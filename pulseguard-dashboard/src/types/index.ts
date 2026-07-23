export interface HostStatus {
  id: string;
  hostname: string;
  ip_address?: string;
  status: string;
  last_seen: string;
  cpu_usage?: number;
  ram_usage?: number;
  disk_usage?: number;
}

export interface ThresholdConfig {
  max_cpu_usage: number;
  max_ram_usage: number;
  max_disk_usage: number;
  error_alert_limit: number;
}

export interface HostDetail extends HostStatus {
  ip_address: string;
  os: string;
  threshold: ThresholdConfig;
}