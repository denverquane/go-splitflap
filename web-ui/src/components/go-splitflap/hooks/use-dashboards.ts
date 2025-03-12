import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

/**
 * Hook to fetch all dashboards
 * 
 * @returns {Object} All dashboard data and dashboard names
 */
export function useDashboards() {
  const result = useQuery({
    queryKey: ['dashboards'],
    queryFn: async () => {
      try {
        const res = await fetch('/api/dashboards');
        
        if (!res.ok) {
          throw new Error(`Failed to fetch dashboards: ${res.status} ${res.statusText}`);
        }
        
        return res.json();
      } catch (error) {
        console.error('Error fetching dashboards:', error);
        throw error;
      }
    },
    staleTime: 30 * 1000, // 30 seconds
  });

  // Extract dashboard names and sort them alphabetically
  const dashboardNames = result.data ? 
    Object.keys(result.data).sort((a, b) => a.localeCompare(b)) : 
    [];
  
  return {
    dashboardsData: result.data || {},
    dashboards: dashboardNames,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Hook to get a specific dashboard from the full dashboards data
 * 
 * @param name The name of the dashboard to fetch
 * @returns The dashboard data along with loading and error states
 */
export function useDashboard(name: string) {
  const { dashboardsData, isLoading, isError, error, refetch } = useDashboards();
  
  // Get the specific dashboard from the full data
  const dashboard = name && dashboardsData ? dashboardsData[name] : null;

  return {
    dashboard,
    isLoading,
    isError,
    error,
    refetch,
    // Flag to indicate if this specific dashboard exists
    exists: !!dashboard
  };
}

/**
 * Hook to fetch the display size
 * 
 * @returns The size of the display along with loading and error states
 */
export function useDisplaySize() {
  const result = useQuery({
    queryKey: ['displaySize'],
    queryFn: async () => {
      try {
        const res = await fetch('/api/display/size');
        
        if (!res.ok) {
          throw new Error(`Failed to fetch display size: ${res.status} ${res.statusText}`);
        }
        
        return res.json();
      } catch (error) {
        console.error('Error fetching display size:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  return {
    size: result.data,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Hook to activate a dashboard
 * 
 * @returns A mutation function for activating a dashboard
 */
export function useActivateDashboard() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (dashboardName: string) => {
      try {
        const response = await fetch(`/api/dashboards/${dashboardName}/activate`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to activate dashboard: ${errorText || response.statusText}`);
        }
        
        return dashboardName;
      } catch (error) {
        console.error(`Error activating dashboard ${dashboardName}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });
}

/**
 * Hook to delete a dashboard
 * 
 * @returns A mutation function for deleting a dashboard
 */
export function useDeleteDashboard() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (dashboardName: string) => {
      try {
        const response = await fetch(`/api/dashboards/${dashboardName}`, {
          method: 'DELETE',
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to delete dashboard: ${errorText || response.statusText}`);
        }
        
        return dashboardName;
      } catch (error) {
        console.error(`Error deleting dashboard ${dashboardName}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });
}

/**
 * Hook for creating or updating a dashboard
 * 
 * @returns A mutation function for creating or updating a dashboard
 */
export function useCreateOrUpdateDashboard() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ name, dashboard }: { name: string, dashboard: any }) => {
      try {
        const response = await fetch(`/api/dashboards/${name}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(dashboard),
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to create/update dashboard: ${errorText || response.statusText}`);
        }
        
        return { name, dashboard };
      } catch (error) {
        console.error(`Error creating/updating dashboard ${name}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['dashboards'] });
    },
  });
}