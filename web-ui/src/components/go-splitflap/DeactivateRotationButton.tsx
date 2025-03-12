import React, { useState } from 'react';
import { Button } from '@/components/shadcn/ui/button';
import { useDeactivateRotation } from './hooks/use-rotations';
import { StopCircle, Loader2 } from 'lucide-react';
import { toast } from '@/components/shadcn/ui/use-toast';

interface DeactivateRotationButtonProps {
  variant?: 'default' | 'outline' | 'secondary' | 'ghost' | 'destructive';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  fullWidth?: boolean;
}

const DeactivateRotationButton: React.FC<DeactivateRotationButtonProps> = ({
  variant = 'outline',
  size = 'default',
  fullWidth = false,
}) => {
  const [isDeactivated, setIsDeactivated] = useState(false);
  const deactivateMutation = useDeactivateRotation();
  
  const handleDeactivate = async () => {
    try {
      await deactivateMutation.mutateAsync();
      setIsDeactivated(true);
      toast({
        title: 'Rotation Stopped',
        description: 'All rotations have been deactivated.',
        variant: 'default',
      });
      
      // Reset state after 5 seconds
      setTimeout(() => {
        setIsDeactivated(false);
      }, 5000);
    } catch (error) {
      console.error('Failed to deactivate rotation:', error);
      toast({
        title: 'Deactivation Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
    }
  };
  
  const isLoading = deactivateMutation.isPending;
  const isDisabled = isLoading || isDeactivated;
  
  return (
    <Button
      variant={variant}
      size={size}
      disabled={isDisabled}
      onClick={handleDeactivate}
      className={fullWidth ? 'w-full' : ''}
    >
      {isLoading ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          Stopping...
        </>
      ) : isDeactivated ? (
        'Stopped'
      ) : (
        <>
          <StopCircle className="mr-2 h-4 w-4" />
          Stop All Rotations
        </>
      )}
    </Button>
  );
};

export default DeactivateRotationButton;