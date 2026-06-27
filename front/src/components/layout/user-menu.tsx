'use client';

import { LogOut, Settings, Shield } from 'lucide-react';
import { useParams, useRouter } from 'next/navigation';
import { useTranslations } from 'next-intl';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { authAdapter } from '@/shared/auth/auth-adapter';

export function UserMenu() {
  const t = useTranslations();
  const params = useParams();
  const router = useRouter();
  const locale = (params?.locale as string) || 'fa';
  const session = authAdapter.getSession();
  const userName = session.userId || 'User';
  const userEmail = session.userId ? `${session.userId}@notifier.local` : '';
  const userRole = session.roles[0] || 'user';
  const initials = (userName.split(' ').length > 1 ? userName.split(' ').map(n => n[0]).join('') : userName.charAt(0)).toUpperCase().slice(0, 2) || 'U';

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-2">
          <Avatar className="h-6 w-6">
            <AvatarFallback className="text-xs">{initials}</AvatarFallback>
          </Avatar>
          <span className="hidden md:inline text-sm max-w-[100px] truncate">{userName}</span>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        <DropdownMenuLabel>
          <div className="flex flex-col">
            <span>{userName}</span>
            <span className="text-xs font-normal text-muted-foreground">{userEmail}</span>
            <span className="text-xs font-normal text-muted-foreground flex items-center gap-1 mt-1">
              <Shield className="h-3 w-3" />
              {t(`settings.${userRole}`)}
            </span>
          </div>
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => router.push(`/${locale}/settings`)}>
          <Settings className="ml-2 h-4 w-4" />
          {t('navigation.settings')}
        </DropdownMenuItem>
        <DropdownMenuSeparator />
        <DropdownMenuItem disabled>
          <LogOut className="ml-2 h-4 w-4" />
          {t('common.close')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
