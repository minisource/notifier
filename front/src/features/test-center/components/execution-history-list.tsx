'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Badge } from '@minisource/ui';
import { History, Trash2, Clock, CheckCircle2, XCircle } from 'lucide-react';
import { ExecutionHistoryItem } from '../types';

interface ExecutionHistoryListProps {
  history: ExecutionHistoryItem[];
  onSelect: (item: ExecutionHistoryItem) => void;
  onClear: () => void;
  selectedId?: string;
}

export function ExecutionHistoryList({ history, onSelect, onClear, selectedId }: ExecutionHistoryListProps) {
  return (
    <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
      <CardHeader className="py-3 px-4 border-b border-border/50 flex flex-row items-center justify-between">
        <div>
          <CardTitle className="text-sm font-semibold flex items-center gap-2">
            <History className="h-4 w-4 text-primary" /> Execution Audit History
          </CardTitle>
          <CardDescription className="text-xs">
            In-memory execution trail for diagnostic sessions ({history.length} records)
          </CardDescription>
        </div>

        {history.length > 0 && (
          <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:bg-destructive/10 gap-1" onClick={onClear}>
            <Trash2 className="h-3 w-3" /> Clear History
          </Button>
        )}
      </CardHeader>

      <CardContent className="p-0 divide-y divide-border/40 max-h-[500px] overflow-y-auto scrollbar-thin">
        {history.length === 0 ? (
          <div className="p-8 text-center text-xs text-muted-foreground">
            No execution history yet. Run an individual endpoint test or multi-step flow.
          </div>
        ) : (
          history.map((item) => {
            const isSelected = item.id === selectedId;
            return (
              <div
                key={item.id}
                onClick={() => onSelect(item)}
                className={`p-3 text-xs flex items-center justify-between cursor-pointer transition-colors ${
                  isSelected ? 'bg-primary/10 border-l-2 border-primary' : 'hover:bg-muted/30'
                }`}
              >
                <div className="flex items-center gap-3">
                  {item.success ? (
                    <CheckCircle2 className="h-4 w-4 text-emerald-500 shrink-0" />
                  ) : (
                    <XCircle className="h-4 w-4 text-rose-500 shrink-0" />
                  )}
                  <div>
                    <div className="flex items-center gap-2 font-medium">
                      <span>{item.operationName}</span>
                      <Badge variant="outline" className="font-mono text-[10px] px-1.5">
                        {item.request.method}
                      </Badge>
                    </div>
                    <span className="text-[11px] text-muted-foreground font-mono">{item.request.url}</span>
                  </div>
                </div>

                <div className="flex items-center gap-3 text-right">
                  <div>
                    <span className={`font-semibold block ${item.success ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-500'}`}>
                      HTTP {item.statusCode}
                    </span>
                    <span className="text-[10px] text-muted-foreground flex items-center justify-end gap-1">
                      <Clock className="h-3 w-3" /> {item.duration}ms
                    </span>
                  </div>
                  <span className="text-[10px] text-muted-foreground font-mono">{item.timestamp}</span>
                </div>
              </div>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}
