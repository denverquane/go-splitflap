import React, { useState } from 'react';
import { Button } from '@/components/shadcn/ui/button';
import { 
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/shadcn/ui/alert-dialog';
import { useDeleteRotation } from './hooks/use-rotations';
import { Trash2, Loader2 } from 'lucide-react';
import { toast } from '@/components/shadcn/ui/use-toast';

interface DeleteRotationButtonProps {
  rotationName: string;
  variant?: 'default' | 'outline' | 'secondary' | 'ghost' | 'destructive';
  size?: 'default' | 'sm' | 'lg' | 'icon';
  fullWidth?: boolean;
}

const DeleteRotationButton: React.FC<DeleteRotationButtonProps> = ({
  rotationName,
  variant = 'destructive',
  size = 'default',
  fullWidth = false,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const deleteMutation = useDeleteRotation();
  
  const handleDelete = async () => {
    try {
      await deleteMutation.mutateAsync(rotationName);
      setIsOpen(false);
      toast({
        title: 'Rotation Deleted',
        description: `${rotationName} has been deleted.`,
        variant: 'default',
      });
    } catch (error) {
      console.error('Failed to delete rotation:', error);
      toast({
        title: 'Deletion Failed',
        description: error instanceof Error ? error.message : 'Unknown error occurred',
        variant: 'destructive',
      });
    }
  };
  
  const isLoading = deleteMutation.isPending;
  
  return (
    <AlertDialog open={isOpen} onOpenChange={setIsOpen}>
      <AlertDialogTrigger asChild>
        <Button
          variant={variant}
          size={size}
          className={fullWidth ? 'w-full' : ''}
        >
          <Trash2 className="h-4 w-4 mr-2" />
          Delete
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Rotation</AlertDialogTitle>
          <AlertDialogDescription>
            Are you sure you want to delete the rotation "{rotationName}"? This action cannot be undone.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction 
            onClick={handleDelete}
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete'
            )}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
};

export default DeleteRotationButton;