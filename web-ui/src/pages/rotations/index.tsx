import React, { useEffect } from 'react';
import { Link } from 'react-router';
import { Button } from '@/components/shadcn/ui/button';
import { useRotations, useDashboards, useActiveState } from '@/components/go-splitflap/hooks';
import RotationCard from '@/components/go-splitflap/RotationCard';
import DeactivateRotationButton from '@/components/go-splitflap/DeactivateRotationButton';

export default function RotationsPage() {
  const { rotations, rotationsData, isLoading: rotationsLoading, isError: rotationsError, error: rotationsErrorData, refetch: refetchRotations } = useRotations();
  const { dashboardsData, isLoading: dashboardsLoading } = useDashboards();
  const { activeRotation } = useActiveState();

  useEffect(() => {
    // Fetch rotations when component mounts
    refetchRotations();
    
    // Debug information in console
    console.log('Rotations route rendered');
  }, [refetchRotations]);

  // Debug information
  useEffect(() => {
    if (rotationsError) {
      console.error('Rotations error:', rotationsErrorData);
    }
    if (!rotationsLoading && rotationsData) {
      console.log('Rotations data:', rotationsData);
    }
  }, [rotationsData, rotationsLoading, rotationsError, rotationsErrorData]);

  const isLoading = rotationsLoading || dashboardsLoading;
  const isError = rotationsError;
  const errorMessage = rotationsErrorData?.message;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-destructive">
        <h2 className="text-xl font-bold mb-2">Error loading data</h2>
        <p>{errorMessage || 'An unknown error occurred'}</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold">Rotations</h1>
          <p className="text-muted-foreground mt-1">
            Automatically cycle between dashboards
          </p>
        </div>
        <div className="flex gap-2">
          <DeactivateRotationButton />
          <Button>
            <Link to="/rotations/new">Create New Rotation</Link>
          </Button>
        </div>
      </div>

      {rotations.length === 0 ? (
        <div className="text-center py-12">
          <h2 className="text-xl font-medium mb-4">No rotations found</h2>
          <p className="text-muted-foreground mb-6">
            Get started by creating your first rotation.
          </p>
          <Button>
            <Link to="/rotations/new">Create Rotation</Link>
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {rotations.map((name) => (
            <RotationCard 
              key={name} 
              name={name} 
              rotation={rotationsData[name]}
              isActive={name === activeRotation}
            />
          ))}
        </div>
      )}
    </div>
  );
}