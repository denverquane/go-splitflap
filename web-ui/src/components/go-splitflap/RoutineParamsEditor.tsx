import React, { useState, useEffect } from 'react';
import { Parameter } from '@/models/Routine';
import { Input } from '@/components/shadcn/ui/input';
import { Label } from '@/components/shadcn/ui/label';
import { Textarea } from '@/components/shadcn/ui/textarea';
import { Switch } from '@/components/shadcn/ui/switch';
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/shadcn/ui/select';
import { HelpCircle } from 'lucide-react';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/shadcn/ui/tooltip';

interface RoutineParamsEditorProps {
  parameters: Parameter[];
  config: Record<string, any>;
  onChange: (newConfig: Record<string, any>) => void;
  className?: string;
}

const RoutineParamsEditor: React.FC<RoutineParamsEditorProps> = ({
  parameters,
  config,
  onChange,
  className = ''
}) => {
  const [paramValues, setParamValues] = useState<Record<string, any>>(config || {});
  
  // Keep local state in sync with parent's config
  useEffect(() => {
    // Only update if config actually changed and is different from current state
    if (config && JSON.stringify(paramValues) !== JSON.stringify(config)) {
      setParamValues(config);
    }
  // Intentionally omitting paramValues from the dependency array to prevent infinite loops
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [JSON.stringify(config)]);

  // Safely call parent's onChange handler without triggering loops
  const handleChange = (newValues: Record<string, any>) => {
    // Only call onChange if values are different from config
    if (JSON.stringify(newValues) !== JSON.stringify(config)) {
      onChange(newValues);
    }
  };

  // Call handleChange when paramValues changes
  // We're not using useEffect here to avoid dependency issues
  // and potential infinite loops

  // Handle parameter value change
  const handleParamChange = (field: string, value: any) => {
    // Batch the state update and onChange call
    const newValues = {
      ...paramValues,
      [field]: value
    };
    
    // Update local state
    setParamValues(newValues);
    
    // Notify parent (with safety check)
    handleChange(newValues);
  };

  // Extract options from description field if it contains options in format "Options: a,b,c"
  const extractOptions = (description: string): string[] => {
    const optionsMatch = description.match(/options:\s*([^.]+)/i);
    if (optionsMatch && optionsMatch[1]) {
      return optionsMatch[1].split(',').map(o => o.trim());
    }
    return [];
  };

  // Render appropriate input based on parameter type
  const renderInput = (param: Parameter) => {
    const { field, type, name, description } = param;
    const value = paramValues[field] ?? '';
    const options = extractOptions(description);
    
    switch (type.toLowerCase()) {
      case 'bool':
        return (
          <div className="flex items-center justify-between" key={field}>
            <div className="flex items-center">
              <Label htmlFor={field} className="text-xs mr-1">{name}</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs">{description}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Switch
              id={field}
              checked={!!value}
              onCheckedChange={(checked) => handleParamChange(field, checked)}
            />
          </div>
        );
        
      // Numeric fields
      case 'int':
      case 'int64':
      case 'float':
      case 'float64':
        return (
          <div className="space-y-1" key={field}>
            <div className="flex items-center">
              <Label htmlFor={field} className="text-xs">{name}</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs">{description}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Input
              id={field}
              type="number"
              value={value}
              onChange={(e) => {
                const numValue = type.includes('int') ? 
                  parseInt(e.target.value) : 
                  parseFloat(e.target.value);
                handleParamChange(field, isNaN(numValue) ? 0 : numValue);
              }}
              className="h-7 text-xs"
              placeholder={description}
              step={type.includes('int') ? 1 : 0.1}
            />
          </div>
        );

      // String array (comma-separated)
      case 'string[]':
        return (
          <div className="space-y-1" key={field}>
            <div className="flex items-center">
              <Label htmlFor={field} className="text-xs">{name}</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs">{description}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Textarea
              id={field}
              value={Array.isArray(value) ? value.join(', ') : value}
              onChange={(e) => {
                const arrayValue = e.target.value.split(',').map(item => item.trim());
                handleParamChange(field, arrayValue);
              }}
              className="text-xs min-h-[60px] max-h-[120px]"
              placeholder="Comma-separated values"
            />
          </div>
        );
        
      // Dropdown for fields with options  
      case 'string':
        if (options.length > 0) {
          return (
            <div className="space-y-1" key={field}>
              <div className="flex items-center">
                <Label htmlFor={field} className="text-xs">{name}</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="text-xs">{description}</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <Select 
                value={String(value)} 
                onValueChange={(val) => handleParamChange(field, val)}
              >
                <SelectTrigger className="h-7 text-xs">
                  <SelectValue placeholder="Select..." />
                </SelectTrigger>
                <SelectContent>
                  {options.map(option => (
                    <SelectItem key={option} value={option}>{option}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          );
        }
        
        // Regular text input (with textarea for multiline)
        if (description.toLowerCase().includes('multiline') || field.toLowerCase().includes('text')) {
          return (
            <div className="space-y-1" key={field}>
              <div className="flex items-center">
                <Label htmlFor={field} className="text-xs">{name}</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="text-xs">{description}</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <Textarea
                id={field}
                value={value}
                onChange={(e) => handleParamChange(field, e.target.value)}
                className="text-xs min-h-[60px] max-h-[120px]"
                placeholder={description}
              />
            </div>
          );
        }
        
        // Standard string input
        return (
          <div className="space-y-1" key={field}>
            <div className="flex items-center">
              <Label htmlFor={field} className="text-xs">{name}</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs">{description}</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Input
              id={field}
              value={value}
              onChange={(e) => handleParamChange(field, e.target.value)}
              className="h-7 text-xs"
              placeholder={description}
            />
          </div>
        );
        
      default:
        // Fallback for any other type
        return (
          <div className="space-y-1" key={field}>
            <div className="flex items-center">
              <Label htmlFor={field} className="text-xs">{name}</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <HelpCircle className="h-3 w-3 ml-1 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p className="text-xs">{description} (Type: {type})</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Input
              id={field}
              value={value}
              onChange={(e) => handleParamChange(field, e.target.value)}
              className="h-7 text-xs"
              placeholder={`Enter ${name}`}
            />
          </div>
        );
    }
  };

  // If no parameters, show a message
  if (!parameters || parameters.length === 0) {
    return (
      <div className="text-xs text-muted-foreground py-2">
        This routine type has no configurable parameters.
      </div>
    );
  }

  return (
    <div className={`space-y-3 py-1 ${className}`}>
      {parameters.map(param => renderInput(param))}
    </div>
  );
};

export default RoutineParamsEditor;