'use client';

import { useTranslations } from 'next-intl';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@minisource/ui';
import { Button } from '@minisource/ui';
import { MoreHorizontal, Eye, Edit, Wifi, TestTube, CheckCircle, Ban, Star, Trash2, Loader2 } from 'lucide-react';
import type { Provider } from '../types';

interface ProviderActionsMenuProps {
  provider: Provider;
  checking: boolean;
  togglePending: boolean;
  defaultPending: boolean;
  onView: () => void;
  onEdit: () => void;
  onHealthCheck: () => void;
  onTest: () => void;
  onToggleStatus: () => void;
  onSetDefault: () => void;
  onDelete: () => void;
}

export function ProviderActionsMenu({
  provider,
  checking,
  togglePending,
  defaultPending,
  onView,
  onEdit,
  onHealthCheck,
  onTest,
  onToggleStatus,
  onSetDefault,
  onDelete,
}: ProviderActionsMenuProps) {
  const t = useTranslations();
  const isDisabled = provider.status === 'disabled';

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" className="h-8 w-8" aria-label={t('common.actions')}>
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-52">
        <DropdownMenuLabel>{t('providers.actions')}</DropdownMenuLabel>
        <DropdownMenuSeparator />

        <DropdownMenuItem onClick={onView}>
          <Eye className="ml-2 h-4 w-4" />
          {t('common.view_details')}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onEdit}>
          <Edit className="ml-2 h-4 w-4" />
          {t('common.edit')}
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem onClick={onHealthCheck} disabled={checking}>
          {checking ? <Loader2 className="ml-2 h-4 w-4 animate-spin" /> : <Wifi className="ml-2 h-4 w-4" />}
          {t('providers.health_check')}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onTest}>
          <TestTube className="ml-2 h-4 w-4" />
          {t('providers.test')}
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem onClick={onToggleStatus} disabled={togglePending}>
          {isDisabled ? (
            <CheckCircle className="ml-2 h-4 w-4 text-green-500" />
          ) : (
            <Ban className="ml-2 h-4 w-4" />
          )}
          {isDisabled ? t('common.enable') : t('common.disable')}
        </DropdownMenuItem>
        <DropdownMenuItem onClick={onSetDefault} disabled={defaultPending}>
          <Star className={`ml-2 h-4 w-4 ${provider.isDefault ? 'fill-amber-500 text-amber-500' : ''}`} />
          {provider.isDefault ? t('providers.unset_default') : t('providers.set_default')}
        </DropdownMenuItem>

        <DropdownMenuSeparator />

        <DropdownMenuItem onClick={onDelete} className="text-red-600 focus:text-red-600 dark:text-red-400 dark:focus:text-red-400">
          <Trash2 className="ml-2 h-4 w-4" />
          {t('common.delete')}
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
