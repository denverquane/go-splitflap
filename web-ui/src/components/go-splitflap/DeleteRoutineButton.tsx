import React, { useState } from 'react';
import { Button } from '@/components/shadcn/ui/button';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/shadcn/ui/alert-dialog';
import { toast } from '@/components/shadcn/ui/use-toast';
import { Trash2 } from 'lucide-react';
import { useDeleteRoutine } from './hooks/use-routines';
import { VariantProps } from 'class-variance-authority';
import { ButtonProps } from '@/components/shadcn/ui/button';

interface DeleteRoutineButtonProps extends ButtonProps {
  routineType: string;
  fullWidth?: boolean;
}

const DeleteRoutineButton: React.FC<DeleteRoutineButtonProps> = ({ 
  routineType, 
  fullWidth = false,
  variant = "default",
  size = "default",
  ...props 
}) => {
  const [showDialog, setShowDialog] = useState(false);
  const deleteRoutine = useDeleteRoutine();

  const handleDelete = async () => {
    try {
      await deleteRoutine.mutateAsync(routineType);
      
      toast({
        title: "Routine deleted",
        description: `Successfully deleted routine "${routineType}"`,
      });
    } catch (error) {
      toast({
        title: "Error",
        description: error instanceof Error ? error.message : "Failed to delete routine",
        variant: "destructive",
      });
    } finally {
      setShowDialog(false);
    }
  };

  return (
    <>
      <Button 
        onClick={() => setShowDialog(true)} 
        variant={variant}
        size={size}
        className={`${fullWidth ? 'w-full' : ''} text-destructive hover:text-destructive-foreground hover:bg-destructive`}
        {...props}
      >
        <Trash2 className="h-4 w-4 mr-2" />
        Delete
      </Button>

      <AlertDialog open={showDialog} onOpenChange={setShowDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Routine</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the routine "{routineType}"? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
};

export default DeleteRoutineButton;