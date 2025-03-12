import { useQuery } from '@tanstack/react-query';

/**
 * Hook to fetch valid characters for the display
 * 
 * @returns {Object} Set of valid characters and related states
 */
export function useDisplayAlphabet() {
  const result = useQuery({
    queryKey: ['displayAlphabet'],
    queryFn: async () => {
      try {
        const response = await fetch('/api/display/alphabet');
        
        if (!response.ok) {
          throw new Error(`Failed to fetch display alphabet: ${response.status} ${response.statusText}`);
        }
        
        const charCodes: number[] = await response.json();
        const validCharsSet = new Set<string>();
        
        // Always add space as valid
        validCharsSet.add(' ');
        
        // Convert ASCII codes to characters
        charCodes.forEach(code => {
          validCharsSet.add(String.fromCharCode(code));
        });
        
        return validCharsSet;
      } catch (error) {
        console.error('Error fetching valid characters:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes - alphabet rarely changes
  });

  return {
    validChars: result.data || new Set([' ']), // Default to space if no data
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Helper function to group valid characters by category
 * 
 * @param validChars Set of valid characters
 * @returns Categorized character groups
 */
export function getGroupedValidChars(validChars: Set<string>) {
  // Convert the set to an array and sort
  const charsArray = Array.from(validChars).sort();
  
  // Create groups for better organization
  const groups = {
    space: [] as string[],
    digits: [] as string[],
    uppercase: [] as string[],
    lowercase: [] as string[],
    punctuation: [] as string[],
    other: [] as string[],
  };
  
  // Process each character
  charsArray.forEach(char => {
    if (char === ' ') {
      groups.space.push(char);
    } else if (/[0-9]/.test(char)) {
      groups.digits.push(char);
    } else if (/[A-Z]/.test(char)) {
      groups.uppercase.push(char);
    } else if (/[a-z]/.test(char)) {
      groups.lowercase.push(char);
    } else if (/[.,\/#!$%\^&\*;:{}=\-_`~()'"?]/.test(char)) {
      groups.punctuation.push(char);
    } else {
      groups.other.push(char);
    }
  });
  
  return groups;
}

/**
 * Hook to update the display with new text
 * 
 * @returns {Object} Functions for updating the display text
 */
export function useUpdateDisplay() {
  /**
   * Update the display with the provided text
   * 
   * @param text The text to display
   * @returns A promise that resolves when the update is complete
   */
  const updateDisplay = async (text: string) => {
    try {
      const response = await fetch('/api/display/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ text }),
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to update display: ${errorText || response.statusText}`);
      }
      
      return true;
    } catch (error) {
      console.error('Failed to update display:', error);
      throw error;
    }
  };

  /**
   * Clear the display
   * 
   * @returns A promise that resolves when the clear operation is complete
   */
  const clearDisplay = async () => {
    try {
      const response = await fetch('/api/display/clear', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to clear display: ${errorText || response.statusText}`);
      }
      
      return true;
    } catch (error) {
      console.error('Failed to clear display:', error);
      throw error;
    }
  };

  return {
    updateDisplay,
    clearDisplay,
  };
}