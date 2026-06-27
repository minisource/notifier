'use client';

import { create } from 'zustand';

interface UIState {
  isSidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  selectedUserId: string | null;
  setSelectedUserId: (id: string | null) => void;
}

export const useUIStore = create<UIState>()((set) => ({
  isSidebarOpen: true,
  toggleSidebar: () => set((s) => ({ isSidebarOpen: !s.isSidebarOpen })),
  setSidebarOpen: (open) => set({ isSidebarOpen: open }),
  selectedUserId: null,
  setSelectedUserId: (id) => set({ selectedUserId: id }),
}));
