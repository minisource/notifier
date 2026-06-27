'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Button } from '@/components/ui/button';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from '@/components/ui/command';
import { Check, ChevronsUpDown } from 'lucide-react';
import { cn } from '@/lib/utils';
import { ChannelBadge } from '@/components/shared/channel-badge';

interface TemplateOption {
  id: string;
  key?: string;
  name: string;
  type: string;
  locale: string;
}

interface TemplateKeyComboboxProps {
  templates: TemplateOption[];
  value?: string;
  onChange: (templateId: string, templateKey?: string) => void;
  loading?: boolean;
}

export function TemplateKeyCombobox({ templates, value, onChange, loading }: TemplateKeyComboboxProps) {
  const t = useTranslations();
  const [open, setOpen] = useState(false);

  const selected = templates.find(t => t.id === value);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={loading}
          className="w-full justify-between"
        >
          {selected ? (
            <span className="flex items-center gap-2">
              <span>{selected.key || selected.name}</span>
              <ChannelBadge channel={selected.type} size="sm" showIcon={false} />
              <span className="text-xs text-muted-foreground">{selected.locale}</span>
            </span>
          ) : (
            <span className="text-muted-foreground">{t('notifications.form.select_template')}</span>
          )}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[300px] p-0">
        <Command>
          <CommandInput placeholder={t('forms.search_options')} />
          <CommandList>
            <CommandEmpty>{t('templates.no_templates')}</CommandEmpty>
            <CommandGroup>
              {templates.map((template) => (
                <CommandItem
                  key={template.id}
                  value={`${template.key || template.name}-${template.locale}`}
                  onSelect={() => {
                    onChange(template.id, template.key);
                    setOpen(false);
                  }}
                >
                  <Check
                    className={cn(
                      'ml-2 h-4 w-4',
                      value === template.id ? 'opacity-100' : 'opacity-0'
                    )}
                  />
                  <div className="flex flex-col">
                    <span>{template.key || template.name}</span>
                    <span className="text-xs text-muted-foreground">
                      {template.name} · {template.locale === 'fa' ? 'فارسی' : 'English'}
                    </span>
                  </div>
                  <ChannelBadge channel={template.type} size="sm" showIcon={false} className="mr-auto" />
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
