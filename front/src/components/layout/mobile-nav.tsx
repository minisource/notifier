'use client';

import { MobileSidebar } from '@minisource/app-shell';
import { Sidebar } from '@/components/layout/sidebar';

interface MobileNavProps {
  open: boolean;
  onClose: () => void;
}

export function MobileNav({ open, onClose }: MobileNavProps) {
  return (
    <MobileSidebar open={open} onClose={onClose}>
      <Sidebar />
    </MobileSidebar>
  );
}

