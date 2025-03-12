import { useQuery } from '@tanstack/react-query';
import { RoutinesResponseSchema, RoutinesResponse, Parameter } from '@/models/Routine';

/**
 * Hook to fetch all routines
 * 
 * @returns {Object} All routine data and routine names
 */
export function useRoutines() {
  const result = useQuery({
    queryKey: ['routines'],
    queryFn: async () => {
      try {
        const res = await fetch('/api/routines');
        
        if (!res.ok) {
          throw new Error(`Failed to fetch routines: ${res.status} ${res.statusText}`);
        }
        
        const rawData = await res.json();
        console.log('Raw routines data:', rawData);
        
        // Validate the data against our schema
        const validationResult = RoutinesResponseSchema.safeParse(rawData);
        
        if (!validationResult.success) {
          console.error('Validation errors:', validationResult.error.format());
          // Still return the data but log errors
          return rawData as RoutinesResponse;
        }
        
        // Return the validated data
        return validationResult.data;
      } catch (error) {
        console.error('Error fetching routines:', error);
        throw error;
      }
    },
    staleTime: 30 * 1000, // 30 seconds
  });

  // Extract routine names and sort them alphabetically
  const routineNames = result.data ? 
    Object.keys(result.data).sort((a, b) => a.localeCompare(b)) : 
    [];
  
  return {
    routinesData: result.data || {},
    routines: routineNames,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Hook to get a specific routine from the full routines data
 * 
 * @param name The name of the routine to fetch
 * @returns The routine data along with loading and error states
 */
export function useRoutine(name: string) {
  const { routinesData, isLoading, isError, error, refetch } = useRoutines();
  
  // Get the specific routine from the full data
  const routine = name && routinesData ? routinesData[name] : null;

  return {
    routine,
    isLoading,
    isError,
    error,
    refetch,
    // Flag to indicate if this specific routine exists
    exists: !!routine
  };
}