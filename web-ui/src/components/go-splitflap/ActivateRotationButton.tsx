import React, { useState } from 'react';
import { Button } from '@/components/shadcn/ui/button';
import { useActivateRotation } from './hooks/use-rotations';
import { Play, Loader2, CheckCircle2 } from 'lucide-react';
import { toast } from '@/components/shadcn/ui/use-toast';

interface ActivateRotationButtonProps {
  rotationName: string;
  variant?: 'default' | 'outline' | 'secondary' | 'ghost' | 'destructive';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  fullWidth?: boolean;
  isActive?: boolean;
}

const ActivateRotationButton: React.FC<ActivateRotationButtonProps> = ({
  rotationName,
  variant = 'default',
  size = 'default',
  fullWidth = false,
  isActive = false,
}) => {
  const [isActivating, setIsActivating] = useState(false);
  const activateMutation = useActivateRotation();
  
  const handleActivate = async () => {
    if (isActive) return;
    
    try {
      setIsActivating(true);
      await activateMutation.mutateAsync(rotationName);
      toast({
        title: 'Rotation Activated',
        description: `${rotationName} rotation is now running on the display.`,
        variant: 'default',
      });
      
      // Reset activating state after a delay
      setTimeout(() => {
        setIsActivating(false);
      }, 1000);
    } catch (error) {
      console.error('Failed to activate rotation:', error);
      toast({
        title: 'Activation Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
      setIsActivating(false);
    }
  };
  
  const isLoading = activateMutation.isPending || isActivating;
  const isDisabled = isLoading || isActive;
  
  return (
    <Button
      variant={isActive ? 'outline' : variant}
      size={size}
      disabled={isDisabled}
      onClick={handleActivate}
      className={`${fullWidth ? 'w-full' : ''} ${isActive ? 'bg-primary/10 border-primary text-primary cursor-default' : ''}`}
    >
      {isLoading ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          Activating...
        </>
      ) : isActive ? (
        <>
          <CheckCircle2 className="mr-2 h-4 w-4" />
          Currently Active
        </>
      ) : (
        <>
          <Play className="mr-2 h-4 w-4" />
          Activate
        </>
      )}
    </Button>
  );
};

export default ActivateRotationButton;