import { Link } from 'react-router';
import { useDashboards, useDisplaySize } from '@/components/go-splitflap/hooks';
import { Button } from '@/components/shadcn/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { useEffect } from 'react';
import DisplayPreview from '@/components/go-splitflap/DisplayPreview';
import ActivateDashboardButton from '@/components/go-splitflap/ActivateDashboardButton';
import DeleteDashboardButton from '@/components/go-splitflap/DeleteDashboardButton';
import { useActiveState } from '@/components/go-splitflap/hooks';

export default function DashboardsPage() {
  const { dashboards, dashboardsData = {}, isLoading: dashboardsLoading, isError: dashboardsError, error: dashboardsErrorData, refetch: refetchDashboards } = useDashboards();
  const { size, isLoading: sizeLoading, isError: sizeError, error: sizeErrorData } = useDisplaySize();
  const { activeDashboard } = useActiveState();
  
  useEffect(() => {
    // Fetch dashboards when component mounts
    refetchDashboards();
    
    // Debug information in console
    console.log('Dashboards route rendered');
  }, [refetchDashboards]);

  // Debug information
  useEffect(() => {
    if (dashboardsError) {
      console.error('Dashboard error:', dashboardsErrorData);
    }
    if (!dashboardsLoading) {
      console.log('Dashboards data:', dashboards);
    }
    if (sizeError) {
      console.error('Size error:', sizeErrorData);
    }
    if (!sizeLoading && size) {
      console.log('Display size:', size);
    }
  }, [dashboards, dashboardsLoading, dashboardsError, dashboardsErrorData, 
      size, sizeLoading, sizeError, sizeErrorData]);

  const isLoading = dashboardsLoading || sizeLoading;
  const isError = dashboardsError || sizeError;
  const errorMessage = dashboardsErrorData?.message || sizeErrorData?.message;

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
        <h1 className="text-3xl font-bold">Dashboards</h1>
        <Button>
          <Link to="/dashboards/new">Create New Dashboard</Link>
        </Button>
      </div>

      {dashboards.length === 0 ? (
        <div className="text-center py-12">
          <h2 className="text-xl font-medium mb-4">No dashboards found</h2>
          <p className="text-muted-foreground mb-6">
            Get started by creating your first dashboard.
          </p>
          <Button>
            <Link to="/dashboards/new">Create Dashboard</Link>
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {dashboards.map((name) => (
            <Card 
              key={name} 
              className={`overflow-hidden ${name === activeDashboard ? 'border-2 border-primary shadow-lg' : ''}`}
            >
              <CardHeader className={name === activeDashboard ? 'bg-primary/10' : ''}>
                <div className="flex items-center justify-between">
                  <CardTitle>{name}</CardTitle>
                  {name === activeDashboard && (
                    <span className="px-2 py-1 bg-primary text-primary-foreground text-xs font-medium rounded-full">
                      Active
                    </span>
                  )}
                </div>
                <CardDescription>
                  {dashboardsData?.[name]?.routines?.length 
                    ? 'Routines: [' + dashboardsData[name].routines.map(r => r.type).join(', ') + ']'
                    : 'Empty dashboard'
                  }
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="h-36">
                  {size && dashboardsData?.[name] ? (
                    <DisplayPreview 
                      width={size.width} 
                      height={size.height} 
                      routineLocations={dashboardsData[name]?.routines?.map((routine, idx) => ({
                        id: `${name}-routine-${idx}`,
                        type: routine.type,
                        location: routine.location,
                        size: routine.size
                      }))}
                    />
                  ) : (
                    <div className="h-full bg-muted rounded-md flex items-center justify-center">
                      <p className="text-muted-foreground">Display size not available</p>
                    </div>
                  )}
                </div>
              </CardContent>
              <CardFooter className="flex flex-col gap-2">
                <ActivateDashboardButton 
                  dashboardName={name} 
                  fullWidth
                  isActive={name === activeDashboard}
                />
                <div className="flex justify-between w-full gap-2">
                  <Button variant="outline" className="flex-1">
                    <Link to={`/dashboards/${name}/edit`} className="w-full">Edit</Link>
                  </Button>
                  <Button variant="secondary" className="flex-1">
                    <Link to={`/dashboards/${name}`} className="w-full">View</Link>
                  </Button>
                </div>
                <DeleteDashboardButton 
                  dashboardName={name} 
                  fullWidth
                  variant="ghost" 
                  size="sm"
                />
              </CardFooter>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}