'use client';

import React from 'react';
import { Button, Badge } from '@minisource/ui';
import { AlertTriangle, Send, ShieldAlert, X } from 'lucide-react';

interface ExternalDeliveryModalProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  operationName: string;
  recipient?: string;
  channel?: string;
}

export function ExternalDeliveryModal({
  isOpen,
  onClose,
  onConfirm,
  operationName,
  recipient,
  channel,
}: ExternalDeliveryModalProps) {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-card border border-amber-500/30 rounded-xl max-w-md w-full shadow-2xl overflow-hidden animate-in fade-in zoom-in-95 duration-150">
        <div className="p-4 bg-amber-500/10 border-b border-amber-500/20 flex items-center justify-between">
          <div className="flex items-center gap-2 text-amber-600 dark:text-amber-400 font-bold text-sm">
            <ShieldAlert className="h-5 w-5" />
            <span>External Side-Effect Warning</span>
          </div>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="p-6 space-y-4">
          <div className="space-y-2">
            <h3 className="font-semibold text-base flex items-center gap-2">
              <Send className="h-4 w-4 text-primary" /> {operationName}
            </h3>
            <p className="text-xs text-muted-foreground leading-relaxed">
              Executing this operation may trigger real SMS, Email, or Push notifications to real delivery targets if external provider credentials are active.
            </p>
          </div>

          <div className="p-3 rounded-lg bg-muted/20 border border-border/50 text-xs space-y-1.5 font-mono">
            <div className="flex justify-between">
              <span className="text-muted-foreground">Action Safety Class:</span>
              <Badge variant="outline" className="text-amber-600 border-amber-500/30 bg-amber-500/10">
                EXTERNAL DELIVERY
              </Badge>
            </div>
            {channel && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Channel:</span>
                <span>{channel}</span>
              </div>
            )}
            {recipient && (
              <div className="flex justify-between">
                <span className="text-muted-foreground">Recipient Target:</span>
                <span className="truncate max-w-[200px]">{recipient}</span>
              </div>
            )}
          </div>

          <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-xs text-amber-800 dark:text-amber-300 flex items-start gap-2">
            <AlertTriangle className="h-4 w-4 shrink-0 text-amber-500 mt-0.5" />
            <p>
              Verify that the recipient is a designated test address or sandbox target before confirming.
            </p>
          </div>

          <div className="flex items-center justify-end gap-3 pt-2">
            <Button variant="outline" size="sm" onClick={onClose}>
              Cancel Execution
            </Button>
            <Button
              variant="default"
              size="sm"
              className="bg-amber-600 hover:bg-amber-700 text-white gap-1.5"
              onClick={() => {
                onConfirm();
                onClose();
              }}
            >
              <Send className="h-3.5 w-3.5" /> Authorize & Dispatch
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
