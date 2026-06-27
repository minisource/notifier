'use client';

import { Button } from '@/components/ui/button';
import { Menu, LogOut } from 'lucide-react';
import { useAuthStore } from '@/stores';

export function Header({ onMenuClick }: { onMenuClick: () => void }) {
  const { clearAuth } = useAuthStore();

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center gap-4 border-b bg-background/95 px-6 backdrop-blur">
      <Button variant="ghost" size="icon" className="lg:hidden" onClick={onMenuClick}>
        <Menu className="h-5 w-5" />
      </Button>
      <div className="flex-1" />
      <Button variant="ghost" size="sm" onClick={() => { clearAuth(); }}>
        <LogOut className="mr-2 h-4 w-4" />
        <span className="hidden sm:inline">Logout</span>
      </Button>
    </header>
  );
}
