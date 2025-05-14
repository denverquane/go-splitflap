import React, { useState, useEffect } from 'react';
import { useDisplayWebSocket } from './hooks';
import { useDisplaySize } from './hooks/use-dashboards';
import { useDisplayAlphabet, useUpdateDisplay } from './hooks/use-display-alphabet';
import { useDisplayTranslations } from './hooks/use-display-translations';
import { Badge } from '@/components/shadcn/ui/badge';
import { Button } from '@/components/shadcn/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { RotateCw, Monitor, Signal, SignalZero, Wifi, WifiOff, XCircle, Loader2, Grid, SendHorizonal, Info, Type } from 'lucide-react';
import { 
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/shadcn/ui/popover";
import { toast } from '@/components/shadcn/ui/use-toast';

/**
 * Component that displays the current status of the SplitFlap display
 * using WebSocket for real-time updates
 */
const DisplayStatus: React.FC = () => {
  const { isConnected, displayState, error } = useDisplayWebSocket();
  const { size } = useDisplaySize();
  const { validChars, isLoading: isLoadingChars } = useDisplayAlphabet();
  const { updateDisplay, clearDisplay } = useUpdateDisplay();
  const { 
    translateChar, 
    hasTranslation, 
    reverseTranslateChar,
    hasReverseTranslation
  } = useDisplayTranslations();
  const [isClearing, setIsClearing] = useState(false);
  const [isSending, setIsSending] = useState(false);
  const [editableText, setEditableText] = useState<string[]>([]);
  const [originalText, setOriginalText] = useState<string>('');
  const [editedIndices, setEditedIndices] = useState<Set<number>>(new Set());
  const [isInitialLoad, setIsInitialLoad] = useState(true);
  const [activeCell, setActiveCell] = useState<number | null>(null);

  // Format the time string to be more readable
  const formatTime = (timeString: string | undefined): string => {
    if (!timeString) return '';
    const date = new Date(timeString);
    return date.toLocaleTimeString();
  };
  
  // Update editable text when display state changes
  useEffect(() => {
    if (displayState && size) {
      const displayText = displayState.displayState || displayState.state || '';
      setOriginalText(displayText);
      
      // On first load, initialize the editable text array
      if (isInitialLoad) {
        const chars = [];
        for (let i = 0; i < size.width * size.height; i++) {
          // Apply reverse translations - convert 'd' back to '°', etc.
          const originalChar = i < displayText.length ? displayText[i] : ' ';
          const displayChar = hasReverseTranslation(originalChar) ? 
            reverseTranslateChar(originalChar) : originalChar;
          chars.push(displayChar);
        }
        setEditableText(chars);
        setIsInitialLoad(false);
      } else {
        // For subsequent updates, only update cells that haven't been edited by the user
        setEditableText(prev => {
          const updatedText = [...prev];
          for (let i = 0; i < size.width * size.height; i++) {
            if (!editedIndices.has(i)) {
              const originalChar = i < displayText.length ? displayText[i] : ' ';
              const displayChar = hasReverseTranslation(originalChar) ? 
                reverseTranslateChar(originalChar) : originalChar;
              updatedText[i] = displayChar;
            }
          }
          return updatedText;
        });
      }
    }
  // Remove the hook functions from the dependency array to avoid infinite loops
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [displayState, size, editedIndices, isInitialLoad]);
  
  // Function to discard all user changes
  const handleDiscardChanges = () => {
    if (displayState && size) {
      const displayText = displayState.displayState || displayState.state || '';
      const chars = [];
      for (let i = 0; i < size.width * size.height; i++) {
        const originalChar = i < displayText.length ? displayText[i] : ' ';
        const displayChar = hasReverseTranslation(originalChar) ? 
          reverseTranslateChar(originalChar) : originalChar;
        chars.push(displayChar);
      }
      setEditableText(chars);
      setEditedIndices(new Set());
      
      toast({
        title: 'Changes Discarded',
        description: 'Your edits have been discarded.',
        variant: 'default',
      });
    }
  };
  
  // Handle clear display action
  const handleClearDisplay = async () => {
    try {
      setIsClearing(true);
      await clearDisplay();
      
      toast({
        title: 'Display Cleared',
        description: 'The display has been cleared successfully.',
        variant: 'default',
      });
    } catch (error) {
      console.error('Failed to clear display:', error);
      toast({
        title: 'Clear Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
    } finally {
      setIsClearing(false);
    }
  };
  
  // Handle updating the display with edited text
  const handleUpdateDisplay = async () => {
    try {
      setIsSending(true);
      
      // Convert any display characters ('°') back to their expected backend format ('d')
      const translatedChars = editableText.map(char => hasTranslation(char) ? translateChar(char) : char);
      
      // Join the array of characters into a string
      const newDisplayText = translatedChars.join('').replace(/\u00A0/g, ' ');
      
      await updateDisplay(newDisplayText);
      
      // Update was successful, so clear the edited indices
      setEditedIndices(new Set());
      // Update the original text to match the current edited text (but use the backend version)
      setOriginalText(newDisplayText);
      
      toast({
        title: 'Display Updated',
        description: 'The display text has been updated successfully.',
        variant: 'default',
      });
    } catch (error) {
      console.error('Failed to update display:', error);
      toast({
        title: 'Update Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
    } finally {
      setIsSending(false);
    }
  };
  
  // Handle character change for a specific cell
  const handleCellChange = (index: number, newChar: string) => {
    // Validate the character against the set of valid characters
    // Allow empty input (will be converted to space), valid characters, or characters that translate to valid ones
    const isValidChar = newChar === '' || // Empty is allowed
                       validChars.has(newChar) || // Direct valid char
                       (hasTranslation(newChar) && validChars.has(translateChar(newChar))); // Char translates to valid
    
    const validatedChar = newChar === '' ? ' ' : isValidChar ? newChar : null;
    
    // If the character is invalid, don't update the cell
    if (validatedChar === null) return;
    
    const updatedText = [...editableText];
    updatedText[index] = validatedChar;
    setEditableText(updatedText);
    
    // Track this index as having been edited by the user
    setEditedIndices(prev => {
      const newSet = new Set(prev);
      newSet.add(index);
      return newSet;
    });
    
    // If a character was typed (not deleted/backspaced), advance to the next cell
    if (newChar !== '') {
      setTimeout(() => {
        if (!size) return;
        
        const { width, height } = size;
        const totalCells = width * height;
        
        // Only advance if not at the last cell
        if (index < totalCells - 1) {
          const nextIndex = index + 1;
          
          // Focus the next input
          const inputs = document.querySelectorAll<HTMLInputElement>('.display-cell-input');
          if (inputs[nextIndex]) {
            inputs[nextIndex].focus();
            // Select the text in the newly focused input
            setTimeout(() => inputs[nextIndex].select(), 0);
          }
        }
      }, 10); // Small timeout to ensure the current change is processed first
    }
  };
  
  // Handle keyboard navigation between cells
  const handleKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (!size) return;
    
    const { width, height } = size;
    const row = Math.floor(index / width);
    const col = index % width;
    
    let nextIndex: number | null = null;
    
    switch (e.key) {
      case 'ArrowRight':
        if (col < width - 1) nextIndex = index + 1;
        break;
      case 'ArrowLeft':
        if (col > 0) nextIndex = index - 1;
        break;
      case 'ArrowDown':
        if (row < height - 1) nextIndex = index + width;
        break;
      case 'ArrowUp':
        if (row > 0) nextIndex = index - width;
        break;
      case 'Tab':
        // Move to next/previous cell without wrap-around
        if (!e.shiftKey) {
          // Only advance if not at the last cell
          if (index < width * height - 1) {
            nextIndex = index + 1;
          }
        } else {
          // Only go back if not at the first cell
          if (index > 0) {
            nextIndex = index - 1;
          }
        }
        e.preventDefault(); // Prevent default tab behavior
        break;
      case 'Backspace':
        // Get the current value of this input
        const currentValue = e.currentTarget.value;
        
        if (currentValue) {
          // If there's content, just let the default backspace behavior clear it
          // handleCellChange will be called by onChange after this
        } else if (index > 0) {
          // If cell is already empty and not the first cell, move to previous cell
          nextIndex = index - 1;
          
          // First select the previous cell
          setTimeout(() => {
            const inputs = document.querySelectorAll<HTMLInputElement>('.display-cell-input');
            if (inputs[nextIndex!]) {
              // We need to clear the content of the previous cell
              handleCellChange(nextIndex!, '');
              // And make sure it's selected for easy typing
              inputs[nextIndex!].select();
            }
          }, 10);
          
          e.preventDefault(); // Prevent default backspace behavior
        }
        break;
    }
    
    // Focus the next input if navigation occurred
    if (nextIndex !== null) {
      const inputs = document.querySelectorAll<HTMLInputElement>('.display-cell-input');
      if (inputs[nextIndex]) {
        inputs[nextIndex].focus();
        // Select the text in the newly focused input
        setTimeout(() => inputs[nextIndex].select(), 0);
      }
    }
  };

  // Function to render the display state grid
  const renderDisplayGrid = () => {
    if (!size) return null;
    
    const { width, height } = size;
    
    // Check if the character is different from the original, accounting for translations
    const isCharChanged = (index: number) => {
      const originalChar = index < originalText.length ? originalText[index] : ' ';
      const currentChar = editableText[index] || ' ';
      
      // If the characters are the same, there's no change
      if (originalChar === currentChar) return false;
      
      // If current character translates to the original, it's not considered changed
      if (hasTranslation(currentChar) && translateChar(currentChar) === originalChar) return false;
      
      // If original character would reverse-translate to current, it's not considered changed
      if (hasReverseTranslation(originalChar) && reverseTranslateChar(originalChar) === currentChar) return false;
      
      // Otherwise, it's changed
      return true;
    };
    
    return (
      <div className="bg-muted rounded-md p-2 w-full overflow-hidden">
        <div 
          className="grid gap-1 bg-background w-full"
          style={{
            gridTemplateColumns: `repeat(${width}, 1fr)`,
            gridTemplateRows: `repeat(${height}, 1fr)`,
            aspectRatio: width / height,
            minHeight: "150px"
          }}
        >
          {editableText.map((char, index) => {
            const isChanged = isCharChanged(index);
            const willTranslate = hasTranslation(char);
            const translatedChar = translateChar(char);
            
            return (
              <div 
                key={index}
                className={`
                  flex items-center justify-center overflow-hidden relative
                  ${activeCell === index ? 'border-accent-foreground border-2 bg-accent/30' : 
                    isChanged ? 'border-primary border-2' : 'border-border border'}
                  cursor-text
                  transition-all duration-150
                `}
                style={{ aspectRatio: '1/1' }}
              >
                <input
                  type="text"
                  maxLength={1}
                  value={char === ' ' ? '' : char}
                  onChange={(e) => handleCellChange(index, e.target.value)}
                  onFocus={(e) => {
                    setActiveCell(index);
                    // Select all text in the input when focused
                    e.target.select();
                  }}
                  onBlur={() => setActiveCell(null)}
                  onClick={(e) => {
                    // Ensure text is selected even when clicking inside the already focused input
                    e.currentTarget.select();
                  }}
                  onKeyDown={(e) => handleKeyDown(index, e)}
                  className={`
                    display-cell-input
                    w-full h-full text-center text-4xl font-mono bg-transparent
                    focus:outline-none
                    ${isChanged ? 'text-primary' : willTranslate ? 'text-accent-foreground' : 'text-foreground'}
                    ${activeCell === index ? 'bg-accent/30' : ''}
                    ${isLoadingChars || displayState?.activeDashboard ? 'cursor-not-allowed opacity-50' : 'cursor-text'}
                  `}
                  style={{ caretColor: 'transparent' }}
                  aria-label={`Character at position ${index}`}
                  autoComplete="off"
                  disabled={isLoadingChars || !!displayState?.activeDashboard}
                  title={isLoadingChars ? "Loading valid characters..." : 
                         displayState?.activeDashboard ? 
                         "Editing disabled while a dashboard is active" : 
                         willTranslate ? `Will display as: ${translatedChar}` : 
                         "Enter a valid display character"}
                />
                
                {/* Show translation indicator */}
                {willTranslate && (
                  <div className="absolute top-0 right-0 bg-accent text-accent-foreground text-[8px] px-1 rounded-bl-sm">
                    {translatedChar}
                  </div>
                )}
              </div>
            );
          })}
        </div>
        
        {/* Add update and discard buttons if there are meaningful changes (excluding just translations) */}
        {editableText.some((_, index) => isCharChanged(index)) && (
          <div className="mt-2 flex justify-between">
            <Button 
              variant="outline" 
              size="sm"
              onClick={handleDiscardChanges}
              disabled={isSending}
              className="text-xs"
            >
              <XCircle className="h-3 w-3 mr-1" />
              Discard Changes
            </Button>
            
            <Button 
              variant="default" 
              size="sm"
              onClick={handleUpdateDisplay}
              disabled={isSending}
              className="text-xs"
            >
              {isSending ? (
                <>
                  <Loader2 className="h-3 w-3 mr-1 animate-spin" />
                  Updating...
                </>
              ) : (
                <>
                  <SendHorizonal className="h-3 w-3 mr-1" />
                  Update Display
                </>
              )}
            </Button>
          </div>
        )}
      </div>
    );
  };

  return (
    <Card className="overflow-hidden">
      <CardHeader className="p-4 pb-2">
        <div className="flex justify-between items-center">
          <div>
            <CardTitle className="text-lg">Display Status</CardTitle>
            <CardDescription>Real-time display information</CardDescription>
          </div>
          <Badge variant={isConnected ? "success" : "destructive"} className="ml-2">
            {isConnected ? (
              <>
                <Wifi className="h-3 w-3 mr-1" />
                Connected
              </>
            ) : (
              <>
                <WifiOff className="h-3 w-3 mr-1" />
                Disconnected
              </>
            )}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="p-4 pt-0 pb-0">
        {error && (
          <div className="text-sm text-destructive mb-2">
            Error: {error}
          </div>
        )}

        <div className="space-y-4">
          <div className="space-y-2">
            {displayState ? (
              <>
                <div className="flex items-center">
                  <Monitor className="h-4 w-4 mr-2" />
                  <span className="font-medium mr-2">Active Dashboard:</span>
                  {displayState.activeDashboard ? (
                    <Badge variant="outline">{displayState.activeDashboard}</Badge>
                  ) : (
                    <Badge variant="secondary">None</Badge>
                  )}
                </div>

                <div className="flex items-center text-xs text-muted-foreground">
                  <Signal className="h-3 w-3 mr-1" />
                  Last update: {formatTime(displayState.currentTime)}
                </div>
              </>
            ) : (
              <div className="flex flex-col items-center justify-center py-2">
                <SignalZero className="h-6 w-6 text-muted-foreground mb-2" />
                <p className="text-sm text-muted-foreground">
                  {isConnected ? "Waiting for data..." : "Connecting..."}
                </p>
              </div>
            )}
          </div>
          
          {/* Display Grid - Always visible */}
          {size ? (
            <>
              {renderDisplayGrid()}
              <div className="flex flex-col space-y-1 mt-2">
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <span>Display Content</span>
                  <div className="flex items-center gap-2">
                    <Popover>
                      <PopoverTrigger asChild>
                        <Button 
                          variant="ghost" 
                          size="icon" 
                          className="h-5 w-5 text-muted-foreground hover:text-foreground"
                          disabled={isLoadingChars || !!displayState?.activeDashboard}
                          title={displayState?.activeDashboard ? 
                                 "Character selection disabled while a dashboard is active" : 
                                 "Show valid characters"}
                        >
                          <Type className="h-4 w-4" />
                          <span className="sr-only">Show valid characters</span>
                        </Button>
                      </PopoverTrigger>
                      <PopoverContent className="w-80 p-4">
                        <div className="space-y-4">
                          <div>
                            <h4 className="font-medium text-sm flex items-center gap-1">
                              <Info className="h-4 w-4" />
                              Valid Display Characters
                            </h4>
                            <div className="text-xs text-muted-foreground">
                              Only these characters can be displayed on the physical splitflap display.
                              Click any character to use it.
                            </div>
                            
                            {/* All valid characters */}
                            <div className="mt-3">
                              <div className="flex flex-wrap gap-1">
                                {Array.from(validChars).sort().map((char, i) => (
                                  <Button
                                    key={i}
                                    variant="outline"
                                    size="icon"
                                    className={`h-8 w-8 text-base font-mono ${hasTranslation(char) ? 'border-accent text-accent-foreground' : ''}`}
                                    onClick={() => {
                                      // If a cell is active, update it with this character
                                      if (activeCell !== null) {
                                        handleCellChange(activeCell, char);
                                      }
                                    }}
                                    title={char === ' ' ? 'Space' : hasTranslation(char) ? 
                                      `Character: ${char} (translates to ${translateChar(char)})` : 
                                      `Character: ${char}`}
                                  >
                                    {char === ' ' ? '␣' : char}
                                    {hasTranslation(char) && (
                                      <span className="absolute top-0 right-0 text-[8px] bg-accent text-accent-foreground px-1 rounded-bl-sm">
                                        {translateChar(char)}
                                      </span>
                                    )}
                                  </Button>
                                ))}
                              </div>
                            </div>
                          </div>
                          
                          <div>
                            <h4 className="font-medium text-sm flex items-center gap-1">
                              <Info className="h-4 w-4" />
                              Character Translations
                            </h4>
                            <div className="text-xs text-muted-foreground">
                              These characters will be automatically translated when displayed on the physical device.
                            </div>
                            <div className="mt-2 text-xs flex flex-wrap gap-1">
                              <a href="/settings" className="text-accent-foreground underline">
                                Manage translations in settings
                              </a>
                            </div>
                          </div>
                        </div>
                      </PopoverContent>
                    </Popover>
                    <span>{size.width} × {size.height}</span>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground italic flex items-center gap-1">
                  {isLoadingChars ? (
                    <>
                      <Loader2 className="h-3 w-3 animate-spin" />
                      Loading valid characters...
                    </>
                  ) : displayState?.activeDashboard ? (
                    <>
                      <Info className="h-3 w-3" />
                      Cell editing is disabled while a dashboard is active. Deactivate it to edit cells.
                    </>
                  ) : (
                    <>
                      <Info className="h-3 w-3" />
                      Click any cell to edit. Only valid display characters are allowed. Click the 
                      <Type className="h-3 w-3 mx-1" /> icon to see all valid characters.
                    </>
                  )}
                </div>
              </div>
            </>
          ) : (
            <div className="flex flex-col items-center justify-center py-4 my-2">
              <Grid className="h-8 w-8 text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">
                {isConnected ? "Waiting for display size data..." : "Connecting..."}
              </p>
            </div>
          )}
        </div>
      </CardContent>
      
      <CardFooter className="p-4 pt-2">
        <div className="w-full">
          <Button 
            variant="destructive" 
            size="sm" 
            className="w-full"
            onClick={handleClearDisplay}
            disabled={isClearing || !isConnected}
          >
            {isClearing ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Clearing...
              </>
            ) : (
              <>
                <XCircle className="h-4 w-4 mr-2" />
                {displayState?.activeDashboard ? 
                  "Clear Display and Deactivate Dashboard" : 
                  "Clear Display"}
              </>
            )}
          </Button>
        </div>
      </CardFooter>
    </Card>
  );
};

export default DisplayStatus;