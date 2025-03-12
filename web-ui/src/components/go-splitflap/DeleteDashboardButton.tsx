import React, { useState } from 'react';
import { Button } from '@/components/shadcn/ui/button';
import { useDeleteDashboard } from './hooks/use-dashboards';
import { Trash2, Loader2 } from 'lucide-react';
import { toast } from '@/components/shadcn/ui/use-toast';
import { useNavigate } from 'react-router';

interface DeleteDashboardButtonProps {
  dashboardName: string;
  variant?: 'default' | 'outline' | 'secondary' | 'destructive' | 'ghost';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  fullWidth?: boolean;
  onDelete?: () => void;
  redirectTo?: string;
}

const DeleteDashboardButton: React.FC<DeleteDashboardButtonProps> = ({
  dashboardName,
  variant = 'outline',
  size = 'default',
  fullWidth = false,
  onDelete,
  redirectTo,
}) => {
  const navigate = useNavigate();
  const [isConfirming, setIsConfirming] = useState(false);
  const deleteMutation = useDeleteDashboard();
  
  const handleDeleteClick = async () => {
    if (!isConfirming) {
      setIsConfirming(true);
      return;
    }
    
    try {
      await deleteMutation.mutateAsync(dashboardName);
      toast({
        title: 'Dashboard Deleted',
        description: `${dashboardName} has been deleted successfully.`,
        variant: 'default',
      });
      
      if (onDelete) {
        onDelete();
      }
      
      if (redirectTo) {
        navigate(redirectTo);
      }
    } catch (error) {
      console.error('Failed to delete dashboard:', error);
      toast({
        title: 'Deletion Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
      setIsConfirming(false);
    }
  };
  
  const handleBlur = () => {
    // Reset confirming state when button loses focus
    setTimeout(() => {
      setIsConfirming(false);
    }, 200);
  };
  
  const isLoading = deleteMutation.isPending;
  
  return (
    <Button
      variant={isConfirming ? 'destructive' : variant}
      size={size}
      disabled={isLoading}
      onClick={handleDeleteClick}
      onBlur={handleBlur}
      className={fullWidth ? 'w-full' : ''}
    >
      {isLoading ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          Deleting...
        </>
      ) : (
        <>
          <Trash2 className="mr-2 h-4 w-4" />
          {isConfirming ? 'Confirm Delete' : 'Delete'}
        </>
      )}
    </Button>
  );
};

export default DeleteDashboardButton;