'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Plus, Trash2, Variable } from 'lucide-react';

interface VariableEntry {
  key: string;
  value: string;
}

interface VariablesEditorProps {
  variables?: Record<string, string>;
  onChange: (variables: Record<string, string>) => void;
}

export function VariablesEditor({ variables = {}, onChange }: VariablesEditorProps) {
  const t = useTranslations();
  const entries: VariableEntry[] = Object.entries(variables).map(([key, value]) => ({ key, value }));
  const [localEntries, setLocalEntries] = useState<VariableEntry[]>(
    entries.length > 0 ? entries : []
  );

  const updateEntries = (newEntries: VariableEntry[]) => {
    setLocalEntries(newEntries);
    const obj: Record<string, string> = {};
    for (const entry of newEntries) {
      if (entry.key.trim()) {
        obj[entry.key.trim()] = entry.value;
      }
    }
    onChange(obj);
  };

  const addEntry = () => {
    updateEntries([...localEntries, { key: '', value: '' }]);
  };

  const removeEntry = (index: number) => {
    updateEntries(localEntries.filter((_, i) => i !== index));
  };

  const updateEntry = (index: number, field: keyof VariableEntry, value: string) => {
    const updated = localEntries.map((entry, i) =>
      i === index ? { ...entry, [field]: value } : entry
    );
    // Check for duplicate keys
    if (field === 'key') {
      const duplicateIndex = updated.findIndex(
        (e, i) => i !== index && e.key === value && value.trim() !== ''
      );
      if (duplicateIndex !== -1) return;
    }
    updateEntries(updated);
  };

  return (
    <div className="space-y-2">
      {localEntries.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <Variable className="h-8 w-8 text-muted-foreground/50" />
          <p className="mt-2 text-sm text-muted-foreground">{t('notifications.form.no_variables')}</p>
          <Button variant="outline" size="sm" onClick={addEntry} className="mt-2">
            <Plus className="ml-1.5 h-3.5 w-3.5" />
            {t('notifications.form.add_variable')}
          </Button>
        </div>
      ) : (
        <>
          <div className="space-y-1.5">
            {localEntries.map((entry, index) => (
              <div key={index} className="flex items-center gap-1.5">
                <Input
                  placeholder={t('notifications.form.variable_key')}
                  value={entry.key}
                  onChange={(e) => updateEntry(index, 'key', e.target.value)}
                  className="h-8 w-[140px] text-xs font-mono"
                />
                <Input
                  placeholder={t('notifications.form.variable_value')}
                  value={entry.value}
                  onChange={(e) => updateEntry(index, 'value', e.target.value)}
                  className="h-8 flex-1 text-xs"
                />
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 shrink-0"
                  onClick={() => removeEntry(index)}
                >
                  <Trash2 className="h-3.5 w-3.5 text-muted-foreground hover:text-destructive" />
                </Button>
              </div>
            ))}
          </div>
          <Button variant="outline" size="sm" onClick={addEntry} className="mt-1">
            <Plus className="ml-1.5 h-3.5 w-3.5" />
            {t('notifications.form.add_variable')}
          </Button>
        </>
      )}
    </div>
  );
}
