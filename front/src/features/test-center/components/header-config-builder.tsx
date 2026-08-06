'use client';

import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription, Button, Input, Label } from '@minisource/ui';
import { Lock, Eye, EyeOff, Plus, Trash2, Globe, Server } from 'lucide-react';
import { AuthMode, HeaderPair } from '../types';

interface HeaderConfigBuilderProps {
  authMode: AuthMode;
  setAuthMode: (mode: AuthMode) => void;
  customToken: string;
  setCustomToken: (token: string) => void;
  headers: HeaderPair[];
  setHeaders: React.Dispatch<React.SetStateAction<HeaderPair[]>>;
  targetMode: 'direct' | 'gateway';
  setTargetMode: (mode: 'direct' | 'gateway') => void;
}

export function HeaderConfigBuilder({
  authMode,
  setAuthMode,
  customToken,
  setCustomToken,
  headers,
  setHeaders,
  targetMode,
  setTargetMode,
}: HeaderConfigBuilderProps) {
  const [showToken, setShowToken] = useState(false);

  const addHeader = () => {
    setHeaders((prev) => [...prev, { key: '', value: '' }]);
  };

  const updateHeader = (index: number, key: string, value: string) => {
    setHeaders((prev) => {
      const copy = [...prev];
      copy[index] = { key, value };
      return copy;
    });
  };

  const removeHeader = (index: number) => {
    setHeaders((prev) => prev.filter((_, i) => i !== index));
  };

  return (
    <Card className="border border-border bg-card/60 backdrop-blur-sm shadow-sm">
      <CardHeader className="py-4 border-b border-border/50">
        <CardTitle className="text-sm font-semibold flex items-center gap-2">
          <Lock className="h-4 w-4 text-primary" /> Request Context & Routing Configuration
        </CardTitle>
        <CardDescription className="text-xs">
          Configure authentication credentials, target routing path, and diagnostic request headers.
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4 pt-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Target Routing Mode */}
          <div className="space-y-2">
            <Label className="text-xs font-medium">Target Service Routing</Label>
            <div className="grid grid-cols-2 gap-2">
              <Button
                variant={targetMode === 'direct' ? 'default' : 'outline'}
                size="sm"
                className="h-9 text-xs flex items-center justify-center gap-1.5"
                onClick={() => setTargetMode('direct')}
              >
                <Server className="h-3.5 w-3.5" /> Direct Backend (:9002)
              </Button>
              <Button
                variant={targetMode === 'gateway' ? 'default' : 'outline'}
                size="sm"
                className="h-9 text-xs flex items-center justify-center gap-1.5"
                onClick={() => setTargetMode('gateway')}
              >
                <Globe className="h-3.5 w-3.5" /> Gateway Proxy (:8080)
              </Button>
            </div>
          </div>

          {/* Authentication Mode */}
          <div className="space-y-2">
            <Label className="text-xs font-medium">Authentication Mode</Label>
            <select
              value={authMode}
              onChange={(e) => setAuthMode(e.target.value as AuthMode)}
              className="w-full h-9 px-3 rounded-lg border border-input bg-background text-xs focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="current-session">Current Admin Session Token</option>
              <option value="bearer-token">Temporary Bearer Token Override</option>
              <option value="none">No Authentication (Public API)</option>
            </select>
          </div>
        </div>

        {/* Temporary Bearer Token Input */}
        {authMode === 'bearer-token' && (
          <div className="space-y-1.5">
            <Label className="text-xs font-medium">Temporary Bearer Token</Label>
            <div className="relative">
              <Input
                type={showToken ? 'text' : 'password'}
                value={customToken}
                onChange={(e) => setCustomToken(e.target.value)}
                placeholder="eyJhbGciOi..."
                className="pr-10 h-9 text-xs"
              />
              <button
                type="button"
                onClick={() => setShowToken(!showToken)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              >
                {showToken ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
              </button>
            </div>
          </div>
        )}

        {/* Custom Permitted Request Headers */}
        <div className="space-y-2 pt-2 border-t border-border/50">
          <div className="flex items-center justify-between">
            <Label className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              Permitted Request Headers
            </Label>
            <Button variant="ghost" size="sm" className="h-7 text-xs px-2 gap-1" onClick={addHeader}>
              <Plus className="h-3 w-3" /> Add Header
            </Button>
          </div>

          <div className="space-y-2">
            {headers.map((h, idx) => (
              <div key={idx} className="flex items-center gap-2">
                <Input
                  placeholder="Header Name (e.g. X-Tenant-ID)"
                  value={h.key}
                  onChange={(e) => updateHeader(idx, e.target.value, h.value)}
                  className="h-8 text-xs font-mono"
                />
                <Input
                  placeholder="Header Value"
                  value={h.value}
                  onChange={(e) => updateHeader(idx, h.key, e.target.value)}
                  className="h-8 text-xs font-mono"
                />
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 shrink-0 text-destructive hover:bg-destructive/10"
                  onClick={() => removeHeader(idx)}
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
