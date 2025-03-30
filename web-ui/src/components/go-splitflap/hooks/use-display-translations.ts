import React from 'react';
import { useQuery } from '@tanstack/react-query';

/**
 * Hook to fetch and manage display character translations
 * 
 * @returns {Object} Object containing translations and utility functions
 */
export function useDisplayTranslations() {
  const result = useQuery({
    queryKey: ['displayTranslations'],
    queryFn: async () => {
      try {
        const response = await fetch('/api/display/translations');
        
        if (!response.ok) {
          throw new Error(`Failed to fetch translations: ${response.status} ${response.statusText}`);
        }
        
        const data = await response.json();
        
        // Convert any numeric values to their corresponding characters
        const formattedData: Record<string, string> = {};
        const reverseData: Record<string, string> = {};
        
        Object.entries(data).forEach(([key, value]) => {
          const sourceChar = typeof key === 'number' ? String.fromCodePoint(key) : key;
          const targetChar = typeof value === 'number' ? String.fromCodePoint(value) : String(value);
          
          // Forward mapping (what backend uses)
          formattedData[sourceChar] = targetChar;
          
          // Reverse mapping (for UI display)
          reverseData[targetChar] = sourceChar;
        });
        
        return { forward: formattedData, reverse: reverseData };
      } catch (error) {
        console.error('Error fetching translations:', error);
        throw error;
      }
    },
    staleTime: 5 * 60 * 1000, // 5 minutes - translations rarely change
  });

  /**
   * Apply translations to a string of text
   * 
   * @param text Text to translate
   * @returns Translated text
   */
  const translateText = React.useCallback((text: string): string => {
    if (!result.data || Object.keys(result.data.forward).length === 0) {
      return text;
    }

    let translatedText = '';
    for (const char of text) {
      translatedText += result.data.forward[char] || char;
    }
    
    return translatedText;
  }, [result.data]);

  /**
   * Get the translated version of a single character
   * 
   * @param char Character to translate
   * @returns Translated character or the original if no translation exists
   */
  const translateChar = React.useCallback((char: string): string => {
    if (!result.data || !char) return char;
    return result.data.forward[char] || char;
  }, [result.data]);

  /**
   * Check if a character has a translation
   * 
   * @param char Character to check
   * @returns Boolean indicating if the character has a translation
   */
  const hasTranslation = React.useCallback((char: string): boolean => {
    if (!result.data || !char) return false;
    return !!result.data.forward[char];
  }, [result.data]);

  /**
   * Check if a character is the result of a translation
   * 
   * @param char The possibly translated character
   * @returns Boolean indicating if this character is a translation result
   */
  // This function is memoized to prevent re-renders
  const hasReverseTranslation = React.useCallback((char: string): boolean => {
    if (!result.data || !char) return false;
    return !!result.data.reverse[char];
  }, [result.data]);
  
  /**
   * Convert a translated character back to its original form
   * This function is memoized to prevent re-renders
   * 
   * @param char The translated character received from the backend
   * @returns The original character before translation
   */
  const reverseTranslateChar = React.useCallback((char: string): string => {
    if (!result.data || !char) return char;
    return result.data.reverse[char] || char;
  }, [result.data]);

  return {
    translations: result.data?.forward || {},
    reverseTranslations: result.data?.reverse || {},
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
    refetch: result.refetch,
    translateText,
    translateChar,
    hasTranslation,
    reverseTranslateChar,
    hasReverseTranslation,
  };
}