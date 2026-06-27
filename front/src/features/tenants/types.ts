export interface Tenant {
  id: string;
  name: string;
  slug: string;
  isActive: boolean;
  enabledChannels: string[];
  monthlyQuota: number;
  usedThisMonth: number;
  createdAt: string;
}
