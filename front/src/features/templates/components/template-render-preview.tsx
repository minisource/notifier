'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import {
  Dialog,  DialogContent, DialogHeader, DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';
import { Play, Loader2, AlertTriangle } from 'lucide-react';
import { toast } from 'sonner';

interface TemplateRenderPreviewProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  templateBody: string;
  templateSubject?: string;
  detectedVariables: string[];
}

export function TemplateRenderPreview({
  open, onOpenChange, templateBody, templateSubject, detectedVariables,
}: TemplateRenderPreviewProps) {
  const t = useTranslations();
  const [variables, setVariables] = useState<Record<string, string>>({});
  const [rendering, setRendering] = useState(false);
  const [renderedBody, setRenderedBody] = useState('');
  const [renderedSubject, setRenderedSubject] = useState('');
  const [activeTab, setActiveTab] = useState('editor');

  // Reset when dialog opens
  const handleOpenChange = (o: boolean) => {
    if (o) {
      const initial: Record<string, string> = {};
      for (const v of detectedVariables) {
        initial[v] = '';
      }
      setVariables(initial);
      setRenderedBody('');
      setRenderedSubject('');
      setActiveTab('editor');
    }
    onOpenChange(o);
  };

  const updateVariable = (key: string, value: string) => {
    setVariables(prev => ({ ...prev, [key]: value }));
  };

  const handleRender = () => {
    setRendering(true);
    // Simulate render
    setTimeout(() => {
      let body = templateBody;
      let subject = templateSubject || '';

      for (const [key, value] of Object.entries(variables)) {
        const regex = new RegExp(`\\{\\{${key}\\}\\}`, 'g');
        body = body.replace(regex, value || `{{${key}}}`);
        subject = subject.replace(regex, value || `{{${key}}}`);
      }

      setRenderedBody(body);
      setRenderedSubject(subject);
      setActiveTab('preview');
      setRendering(false);
      toast.success(t('templates.render_preview'));
    }, 500);
  };

  const missingVariables = detectedVariables.filter(v => !variables[v]?.trim());
  const hasMissing = missingVariables.length > 0;

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Play className="h-4 w-4" />
            {t('templates.render_preview')}
          </DialogTitle>
        </DialogHeader>

        <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
          <TabsList className="w-full">
            <TabsTrigger value="editor" className="flex-1">{t('templates.variables')}</TabsTrigger>
            <TabsTrigger value="preview" className="flex-1" disabled={!renderedBody}>
              {t('templates.preview')}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="editor" className="space-y-4 py-4">
            {/* Detected variables */}
            {detectedVariables.length > 0 ? (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium">
                    {t('templates.variables')}
                    {hasMissing && (
                      <Badge variant="secondary" className="ml-2 text-xs">
                        {missingVariables.length} {t('templates.missing_variables')}
                      </Badge>
                    )}
                  </p>
                </div>
                {detectedVariables.map(v => (
                  <div key={v} className="grid grid-cols-[120px_1fr] items-center gap-3">
                    <div className="flex items-center gap-1.5">
                      <code className="rounded bg-muted px-2 py-0.5 text-xs font-mono">{v}</code>
                    </div>
                    <Input
                      value={variables[v] || ''}
                      onChange={(e) => updateVariable(v, e.target.value)}
                      placeholder={`Enter value for {{${v}}}...`}
                      className={!variables[v]?.trim() ? 'border-amber-300 dark:border-amber-700' : ''}
                    />
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground py-4 text-center">
                {t('templates.no_variables') || 'No variables detected in template'}
              </p>
            )}

            {hasMissing && (
              <div className="flex items-center gap-2 rounded-md border border-amber-200 bg-amber-50 p-3 dark:border-amber-900/50 dark:bg-amber-950/20">
                <AlertTriangle className="h-4 w-4 shrink-0 text-amber-600 dark:text-amber-400" />
                <p className="text-xs text-amber-700 dark:text-amber-300">
                  {t('templates.missing_variables')}: {missingVariables.join(', ')}
                </p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="preview" className="space-y-4 py-4">
            {renderedSubject && (
              <div>
                <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.subject')}</p>
                <div className="rounded-md border bg-card p-3 text-sm">{renderedSubject}</div>
              </div>
            )}
            <div>
              <p className="text-xs font-medium text-muted-foreground mb-1">{t('notifications.body')}</p>
              <div className="rounded-md border bg-card p-3 text-sm whitespace-pre-wrap">{renderedBody}</div>
            </div>
          </TabsContent>
        </Tabs>

        <div className="flex items-center justify-end gap-3 pt-4 border-t">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.close')}
          </Button>
          <Button onClick={handleRender} disabled={rendering}>
            {rendering ? <Loader2 className="ml-1.5 h-4 w-4 animate-spin" /> : <Play className="ml-1.5 h-4 w-4" />}
            {t('templates.render_preview')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
