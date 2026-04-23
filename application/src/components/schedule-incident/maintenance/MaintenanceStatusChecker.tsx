import { MaintenanceItem } from '@/services/types/maintenance.types';

interface MaintenanceStatusCheckerProps {
  maintenanceData: MaintenanceItem[];
  onStatusUpdated: () => void;
}

// Status transitions and notifications are now handled by the service-operation
// backend (maintenance-monitoring goroutine). This component is intentionally
// a no-op to prevent double notifications and race conditions.
export const MaintenanceStatusChecker = (_props: MaintenanceStatusCheckerProps) => {
  return null;
};
