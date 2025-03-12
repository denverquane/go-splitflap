import React, { useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { Badge } from '@/components/shadcn/ui/badge';
import { useRoutines } from '@/components/go-splitflap/hooks';

import { 
  Clock, 
  AlarmClock, 
  MessageSquare, 
  Cloud, 
  Type
} from 'lucide-react';

export default function RoutinesPage() {
  const { routines, routinesData, isLoading, isError, error, refetch } = useRoutines();

  useEffect(() => {
    // Fetch routines when component mounts
    refetch();
    
    // Debug information in console
    console.log('Routines route rendered');
  }, [refetch]);

  // Debug information
  useEffect(() => {
    if (isError) {
      console.error('Routines error:', error);
    }
    if (!isLoading && routinesData) {
      console.log('Routines data:', routinesData);
    }
  }, [routinesData, isLoading, isError, error]);

  // Function to get icon based on routine type
  const getRoutineIcon = (type: string) => {
    switch (type) {
      case 'CLOCK':
        return <Clock className="h-5 w-5" />;
      case 'TIMER':
        return <AlarmClock className="h-5 w-5" />;
      case 'TEXT':
        return <Type className="h-5 w-5" />;
      case 'SLOWTEXT':
        return <MessageSquare className="h-5 w-5" />;
      case 'WEATHER':
        return <Cloud className="h-5 w-5" />;
      default:
        return null;
    }
  };

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
        <p>{error instanceof Error ? error.message : 'An unknown error occurred'}</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold">Routines</h1>
          <p className="text-muted-foreground mt-1">
            Routines are the bottom-level components for your Splitflap display! Add and combine them on Dashboards to create interesting displays!
          </p>
        </div>
      </div>

      {routines.length === 0 ? (
        <div className="text-center py-12">
          <h2 className="text-xl font-medium mb-4">No routines found</h2>
          <p className="text-muted-foreground mb-6">
            No routines are currently available in the system.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {routines.map((name) => {
            const routineType = name;
            const routine = routinesData[routineType];
            return (
              <Card key={routineType} className="overflow-hidden">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center">
                      <span className="mr-2 flex items-center justify-center bg-secondary rounded-full p-1.5">
                        {getRoutineIcon(routineType)}
                      </span>
                      {routineType}
                    </CardTitle>
                    <Badge>{routineType}</Badge>
                  </div>
                  <CardDescription>
                    Configurable routine
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3 text-sm">
                    <p className="text-xs text-muted-foreground mb-2">Configurable Parameters:</p>
                    
                    {routine.parameters && routine.parameters.length > 0 ? (
                      <div className="grid grid-cols-1 gap-1.5">
                        {routine.parameters.map((param, index) => (
                          <div key={index} className="flex flex-col space-y-0.5 mb-2">
                            <div className="flex items-center">
                              <div className="w-2 h-2 bg-primary rounded-full mr-2"></div>
                              <span className="font-medium">{param.name}</span>
                              <span className="ml-auto text-xs px-1.5 py-0.5 rounded bg-secondary text-secondary-foreground">
                                {param.type}
                              </span>
                            </div>
                            {param.description && (
                              <div className="pl-4 text-xs text-muted-foreground">
                                {param.description}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    ) : (
                      <div className="text-sm text-muted-foreground italic">
                        No configurable parameters available
                      </div>
                    )}
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}