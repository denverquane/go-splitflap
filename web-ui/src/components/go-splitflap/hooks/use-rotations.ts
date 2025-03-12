import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Rotation, RotationEntry } from '@/models/Rotation';

/**
 * Hook to fetch all rotations
 * 
 * @returns {Object} All rotation data and rotation names
 */
export function useRotations() {
  const result = useQuery({
    queryKey: ['rotations'],
    queryFn: async () => {
      try {
        const res = await fetch('/api/rotations');
        
        if (!res.ok) {
          throw new Error(`Failed to fetch rotations: ${res.status} ${res.statusText}`);
        }
        
        const data = await res.json();
        console.log('Fetched rotations data:', data);
        return data;
      } catch (error) {
        console.error('Error fetching rotations:', error);
        throw error;
      }
    },
    staleTime: 30 * 1000, // 30 seconds
  });

  // Extract rotation names and sort them alphabetically
  const rotationNames = result.data ? 
    Object.keys(result.data).sort((a, b) => a.localeCompare(b)) : 
    [];
  
  return {
    rotationsData: result.data || {},
    rotations: rotationNames,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Hook to get a specific rotation from the full rotations data
 * 
 * @param name The name of the rotation to fetch
 * @returns The rotation data along with loading and error states
 */
export function useRotation(name: string) {
  const { rotationsData, isLoading, isError, error, refetch } = useRotations();
  
  // Get the specific rotation from the full data
  const rotation = name && rotationsData ? rotationsData[name] : null;

  return {
    rotation,
    isLoading,
    isError,
    error,
    refetch,
    // Flag to indicate if this specific rotation exists
    exists: !!rotation
  };
}

/**
 * Hook to activate a rotation
 * 
 * @returns A mutation function for activating a rotation
 */
export function useActivateRotation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (rotationName: string) => {
      try {
        const response = await fetch(`/api/rotations/${rotationName}/activate`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({}),
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to activate rotation: ${errorText || response.statusText}`);
        }
        
        return rotationName;
      } catch (error) {
        console.error(`Error activating rotation ${rotationName}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['rotations'] });
    },
  });
}

/**
 * Hook to deactivate all rotations
 * 
 * @returns A mutation function for deactivating rotations
 */
export function useDeactivateRotation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      try {
        const response = await fetch(`/api/rotations/deactivate`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({}),
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to deactivate rotation: ${errorText || response.statusText}`);
        }
        
        return true;
      } catch (error) {
        console.error(`Error deactivating rotation:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['rotations'] });
    },
  });
}

/**
 * Hook to create or update a rotation
 * 
 * @returns A mutation function for creating or updating a rotation
 */
export function useCreateOrUpdateRotation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ name, entries }: { name: string, entries: RotationEntry[] }) => {
      try {
        const rotation: Rotation = {
          rotation: entries
        };
        
        const response = await fetch(`/api/rotations/${name}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(rotation),
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to save rotation: ${errorText || response.statusText}`);
        }
        
        return name;
      } catch (error) {
        console.error(`Error saving rotation ${name}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['rotations'] });
    },
  });
}

/**
 * Hook to delete a rotation
 * 
 * @returns A mutation function for deleting a rotation
 */
export function useDeleteRotation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (rotationName: string) => {
      try {
        const response = await fetch(`/api/rotations/${rotationName}`, {
          method: 'DELETE',
        });
        
        if (!response.ok) {
          const errorText = await response.text();
          throw new Error(`Failed to delete rotation: ${errorText || response.statusText}`);
        }
        
        return rotationName;
      } catch (error) {
        console.error(`Error deleting rotation ${rotationName}:`, error);
        throw error;
      }
    },
    onSuccess: () => {
      // Invalidate related queries to refresh data
      queryClient.invalidateQueries({ queryKey: ['rotations'] });
    },
  });
}