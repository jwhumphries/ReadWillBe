export interface Reading {
  id: number;
  planId: number;
  date: string;
  content: string;
  status: 'pending' | 'completed';
  completedAt?: string;
  plan?: Plan;
}

export interface ManualReading {
  id: string;
  date: string;
  content: string;
}

export interface Plan {
  id: number;
  title: string;
  status: 'active' | 'processing' | 'failed';
  errorMessage?: string;
  readings: Reading[];
}

export interface PlanGroup {
  plan: Plan;
  readings: Reading[];
}

export interface FetchOptions extends RequestInit {
  csrfToken?: string;
}

export interface Notification {
  id: number;
  readingId: number;
  reading: Reading;
  createdAt: string;
}
