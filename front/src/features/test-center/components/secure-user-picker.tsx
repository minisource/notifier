'use client';

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useTranslations } from 'next-intl';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Button, toast } from '@minisource/ui';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Check, ChevronsUpDown, Loader2, X, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { authClient } from '@/shared/auth/auth-sdk-client';
import { useTenantStore } from '@/stores/tenant.store';

export interface UserSelection {
  id: string;
  displayName: string | null;
  username?: string | null;
  emailMasked?: string | null;
  phoneMasked?: string | null;
  status?: string;
}

interface SecureUserPickerProps {
  value: string | null; // Stable canonical user ID
  onChange: (userId: string | null, userSummary?: UserSelection | null) => void;
  tenantId?: string;
  disabled?: boolean;
  required?: boolean;
  placeholder?: string;
}

function maskEmail(email: string): string {
  if (!email) return '';
  const [local, domain] = email.split('@');
  if (!domain) return email;
  if (local.length <= 2) {
    return `${local.charAt(0)}***@${domain}`;
  }
  return `${local.charAt(0)}***${local.charAt(local.length - 1)}@${domain}`;
}

function maskPhone(phone: string): string {
  if (!phone) return '';
  if (phone.length <= 5) return '***';
  return `${phone.slice(0, 3)}******${phone.slice(-4)}`;
}

export function SecureUserPicker({
  value,
  onChange,
  tenantId,
  disabled = false,
  placeholder,
}: SecureUserPickerProps) {
  const t = useTranslations();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [users, setUsers] = useState<UserSelection[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedUser, setSelectedUser] = useState<UserSelection | null>(null);

  const activeTenantId = useTenantStore((state) => state.activeTenant?.id);
  const resolvedTenantId = tenantId || activeTenantId;

  const abortControllerRef = useRef<AbortController | null>(null);
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);
  const prevTenantIdRef = useRef<string | undefined>(resolvedTenantId);

  // Clear selection if tenant ID changes to prevent cross-tenant operations
  useEffect(() => {
    if (prevTenantIdRef.current !== resolvedTenantId) {
      prevTenantIdRef.current = resolvedTenantId;
      setSelectedUser(null);
      setUsers([]);
      onChange(null, null);
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    }
  }, [resolvedTenantId, onChange]);

  // Resolve user details when value changes externally
  useEffect(() => {
    if (!value) {
      setSelectedUser(null);
      return;
    }

    if (selectedUser?.id === value) {
      return;
    }

    let isSubscribed = true;
    const fetchUserDetail = async () => {
      try {
        const res = await authClient.users.getById(value);
        if (res && isSubscribed) {
          const userSummary: UserSelection = {
            id: res.id,
            displayName: res.firstName || res.lastName ? `${res.firstName || ''} ${res.lastName || ''}`.trim() : null,
            username: res.username,
            emailMasked: maskEmail(res.email),
            phoneMasked: res.phone ? maskPhone(res.phone) : null,
            status: res.isActive ? 'active' : 'inactive',
          };
          setSelectedUser(userSummary);
        }
      } catch {
        if (isSubscribed) {
          setSelectedUser({
            id: value,
            displayName: t('common.unknown_user') || 'Unknown User',
            status: 'unknown',
          });
        }
      }
    };

    fetchUserDetail();
    return () => {
      isSubscribed = false;
    };
  }, [value, selectedUser?.id, t]);

  const searchUsers = useCallback(
    async (searchQuery: string) => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }

      setLoading(true);
      const controller = new AbortController();
      abortControllerRef.current = controller;

      try {
        const headers: Record<string, string> = {};
        if (resolvedTenantId && resolvedTenantId !== 'all') {
          headers['X-Tenant-Id'] = resolvedTenantId;
        }

        const params: Record<string, string | number> = {
          page: 1,
          pageSize: searchQuery.trim().length >= 2 ? 15 : 10,
        };

        if (searchQuery.trim().length >= 2) {
          params.search = searchQuery;
        }

        // Recipient search via the Auth SDK — same transport/token/tenant
        // headers as production flows; response already normalized + unwrapped.
        const result = await authClient.users.search(params, {
          headers,
          signal: controller.signal,
        });

        const rawUsers = result?.users || [];
        const formattedUsers = rawUsers.map((u: any) => ({
          id: u.id,
          displayName: u.firstName || u.lastName ? `${u.firstName || ''} ${u.lastName || ''}`.trim() : null,
          username: u.username,
          emailMasked: maskEmail(u.email),
          phoneMasked: u.phone ? maskPhone(u.phone) : null,
          status: u.isActive ? 'active' : 'inactive',
        }));

        setUsers(formattedUsers);
      } catch (err: any) {
        if (err.kind !== 'cancelled') {
          toast.error(t('errors.search_failed') || 'Failed to search users');
        }
      } finally {
        if (abortControllerRef.current === controller) {
          setLoading(false);
        }
      }
    },
    [resolvedTenantId, t]
  );

  // Fetch default users when the picker opens or tenant ID changes
  useEffect(() => {
    if (open) {
      searchUsers(query);
    }
  }, [open, resolvedTenantId, searchUsers]);

  const handleQueryChange = (val: string) => {
    setQuery(val);

    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    debounceTimerRef.current = setTimeout(() => {
      searchUsers(val);
    }, 300);
  };

  useEffect(() => {
    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  return (
    <div className="flex items-center gap-2 w-full">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            disabled={disabled}
            className="w-full justify-between h-9 text-xs"
          >
            {selectedUser ? (
              <span className="flex items-center gap-2 truncate">
                <User className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                <span className="font-semibold text-foreground">
                  {selectedUser.displayName || selectedUser.username || selectedUser.id}
                </span>
                {selectedUser.emailMasked && (
                  <span className="text-muted-foreground text-[10px] truncate max-w-[120px]">
                    ({selectedUser.emailMasked})
                  </span>
                )}
              </span>
            ) : (
              <span className="text-muted-foreground">
                {placeholder || t('forms.search_users') || 'Search users by name/email/phone...'}
              </span>
            )}
            <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[340px] p-0" align="start">
          <Command shouldFilter={false}>
            <CommandInput
              value={query}
              onValueChange={handleQueryChange}
              placeholder={t('forms.type_to_search') || 'Type at least 2 characters...'}
            />
            <CommandList>
              {loading && (
                <div className="flex items-center justify-center p-4 gap-2 text-xs text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Searching Auth registry...
                </div>
              )}
              {!loading && users.length === 0 && (
                <CommandEmpty>
                  {query.trim().length >= 2 
                    ? "No users found matching query." 
                    : "No users available in this tenant context."}
                </CommandEmpty>
              )}
              {!loading && users.length > 0 && (
                <CommandGroup>
                  {users.map((user) => (
                    <CommandItem
                      key={user.id}
                      value={user.id}
                      onSelect={() => {
                        setSelectedUser(user);
                        onChange(user.id, user);
                        setOpen(false);
                      }}
                      className="text-xs"
                    >
                      <Check
                        className={cn(
                          'ml-2 h-3.5 w-3.5 shrink-0',
                          value === user.id ? 'opacity-100' : 'opacity-0'
                        )}
                      />
                      <div className="flex flex-col truncate">
                        <span className="font-semibold">
                          {user.displayName || 'Unnamed User'} ({user.username})
                        </span>
                        <span className="text-[10px] text-muted-foreground">
                          {user.emailMasked || 'No Email'} · {user.phoneMasked || 'No Phone'}
                        </span>
                      </div>
                    </CommandItem>
                  ))}
                </CommandGroup>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {value && (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => {
            setSelectedUser(null);
            onChange(null, null);
          }}
          className="h-8 w-8 text-muted-foreground hover:text-foreground shrink-0"
        >
          <X className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
