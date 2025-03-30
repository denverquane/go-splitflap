import React, { useState, useEffect } from 'react';
import { client } from '@/providers';
import { displayContract } from '@/lib/contract';
import { Button } from '@/components/shadcn/ui/button';
import { Input } from '@/components/shadcn/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { useToast } from '@/components/shadcn/ui/use-toast';
import { XIcon, PlusIcon } from 'lucide-react';

const TranslationSettings: React.FC = () => {
  const { toast } = useToast();
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [newSourceChar, setNewSourceChar] = useState('');
  const [newTargetChar, setNewTargetChar] = useState('');
  const [loading, setLoading] = useState(true);

  // Fetch current translations on component mount
  useEffect(() => {
    const fetchTranslations = async () => {
      try {
        const response = await fetch('/api/display/translations');
        
        if (response.ok) {
          const data = await response.json();
          
          // Convert any numeric values to their corresponding characters
          const formattedData: Record<string, string> = {};
          Object.entries(data).forEach(([key, value]) => {
            const sourceChar = typeof key === 'number' ? String.fromCodePoint(key) : key;
            const targetChar = typeof value === 'number' ? String.fromCodePoint(value) : String(value);
            formattedData[sourceChar] = targetChar;
          });
          
          setTranslations(formattedData);
        } else {
          toast({
            title: 'Error',
            description: 'Failed to fetch translations',
            variant: 'destructive',
          });
        }
      } catch (error) {
        console.error('Error fetching translations:', error);
        toast({
          title: 'Error',
          description: 'Failed to fetch translations',
          variant: 'destructive',
        });
      } finally {
        setLoading(false);
      }
    };

    fetchTranslations();
  }, [toast]);

  // Save all translations
  const saveTranslations = async () => {
    try {
      const response = await fetch('/api/display/translations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(translations),
      });

      if (response.ok) {
        toast({
          title: 'Success',
          description: 'Translations saved successfully',
        });
      } else {
        toast({
          title: 'Error',
          description: 'Failed to save translations',
          variant: 'destructive',
        });
      }
    } catch (error) {
      console.error('Error saving translations:', error);
      toast({
        title: 'Error',
        description: 'Failed to save translations',
        variant: 'destructive',
      });
    }
  };

  // Add a new translation
  const addTranslation = () => {
    if (!newSourceChar || !newTargetChar) {
      toast({
        title: 'Error',
        description: 'Both characters must be provided',
        variant: 'destructive',
      });
      return;
    }

    // Using Array.from to count Unicode code points correctly
    if (Array.from(newSourceChar).length !== 1 || Array.from(newTargetChar).length !== 1) {
      toast({
        title: 'Error',
        description: 'Source and target must be single Unicode characters',
        variant: 'destructive',
      });
      return;
    }

    setTranslations((prev) => ({
      ...prev,
      [newSourceChar]: newTargetChar,
    }));

    setNewSourceChar('');
    setNewTargetChar('');
  };

  // Remove a translation
  const removeTranslation = (sourceChar: string) => {
    setTranslations((prev) => {
      const updated = { ...prev };
      delete updated[sourceChar];
      return updated;
    });
  };

  if (loading) {
    return <div>Loading translations...</div>;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Character Translations</CardTitle>
        <CardDescription>
          Define character mappings that replace characters when displayed on the splitflap
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
            <div className="font-medium">Source Character</div>
            <div className="font-medium">Target Character</div>
            <div></div>
          </div>

          {/* Existing translations */}
          {Object.entries(translations).map(([source, target]) => (
            <div key={source} className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
              <div className="font-mono bg-muted p-2 rounded flex items-center justify-center">
                <span>
                  {source}
                </span>
                {typeof source === 'string' && (
                  <span className="text-xs text-muted-foreground ml-2">
                    {`U+${source.codePointAt(0)?.toString(16).toUpperCase().padStart(4, '0')}`}
                  </span>
                )}
              </div>
              <div className="font-mono bg-muted p-2 rounded flex items-center justify-center">
                <span>
                  {target}
                </span>
                {typeof target === 'string' && (
                  <span className="text-xs text-muted-foreground ml-2">
                    {`U+${target.codePointAt(0)?.toString(16).toUpperCase().padStart(4, '0')}`}
                  </span>
                )}
              </div>
              <Button 
                variant="ghost" 
                size="icon"
                onClick={() => removeTranslation(source)}
              >
                <XIcon className="h-4 w-4" />
              </Button>
            </div>
          ))}

          {/* Add new translation */}
          <div className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center border-t pt-4">
            <div className="relative">
              <Input
                type="text"
                value={newSourceChar}
                onChange={(e) => {
                  // Take only the first Unicode character
                  const chars = Array.from(e.target.value);
                  if (chars.length > 0) {
                    setNewSourceChar(chars[0]);
                  } else {
                    setNewSourceChar('');
                  }
                }}
                placeholder="@"
                className="font-mono text-center text-lg"
              />
              {newSourceChar && typeof newSourceChar === 'string' && (
                <div className="absolute bottom-1 right-2 text-xs text-muted-foreground">
                  {`U+${newSourceChar.codePointAt(0)?.toString(16).toUpperCase().padStart(4, '0')}`}
                </div>
              )}
            </div>
            <div className="relative">
              <Input
                type="text"
                value={newTargetChar}
                onChange={(e) => {
                  // Take only the first Unicode character
                  const chars = Array.from(e.target.value);
                  if (chars.length > 0) {
                    setNewTargetChar(chars[0]);
                  } else {
                    setNewTargetChar('');
                  }
                }}
                placeholder="~"
                className="font-mono text-center text-lg"
              />
              {newTargetChar && typeof newTargetChar === 'string' && (
                <div className="absolute bottom-1 right-2 text-xs text-muted-foreground">
                  {`U+${newTargetChar.codePointAt(0)?.toString(16).toUpperCase().padStart(4, '0')}`}
                </div>
              )}
            </div>
            <Button onClick={addTranslation} size="icon">
              <PlusIcon className="h-4 w-4" />
            </Button>
          </div>

          {/* Save button */}
          <div className="flex justify-end pt-4">
            <Button onClick={saveTranslations}>
              Save Translations
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default TranslationSettings;