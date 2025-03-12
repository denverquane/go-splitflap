import { useParams, Link } from 'react-router';
import { useEffect } from 'react';
import { useDashboard, useDisplaySize } from '@/components/go-splitflap/hooks';
import { Button } from '@/components/shadcn/ui/button';
import DisplayPreview from '@/components/go-splitflap/DisplayPreview';
import ActivateDashboardButton from '@/components/go-splitflap/ActivateDashboardButton';
import DeleteDashboardButton from '@/components/go-splitflap/DeleteDashboardButton';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { ChevronLeft } from 'lucide-react';

const RoutineTypeLabels: Record<string, string> = {
  TEXT: 'Static Text',
  CLOCK: 'Clock',
  TIMER: 'Timer',
  SLOWTEXT: 'Slow Text',
  WEATHER: 'Weather',
};

const RoutineTypeDescriptions: Record<string, string> = {
  TEXT: 'Displays static text on the splitflap',
  CLOCK: 'Shows the current time',
  TIMER: 'Countdown timer',
  SLOWTEXT: 'Displays text with a typing effect',
  WEATHER: 'Shows weather information',
};

export default function DashboardDetailPage() {
  const { name } = useParams<{ name: string }>();
  const { dashboard, isLoading: dashboardLoading, isError: dashboardError, error: dashboardErrorData, refetch: refetchDashboard } = useDashboard(name);
  const { size, isLoading: sizeLoading, isError: sizeError } = useDisplaySize();

  useEffect(() => {
    if (name) {
      refetchDashboard();
    }
    
    // Debug information
    console.log('Dashboard detail page rendered for:', name);
  }, [name, refetchDashboard]);

  useEffect(() => {
    if (dashboardError) {
      console.error('Dashboard error:', dashboardErrorData);
    }
    if (!dashboardLoading && dashboard) {
      console.log('Dashboard data:', dashboard);
    }
  }, [dashboard, dashboardLoading, dashboardError, dashboardErrorData]);

  const isLoading = dashboardLoading || sizeLoading;
  const isError = dashboardError || sizeError;
  const errorMessage = dashboardErrorData?.message;

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
        <h2 className="text-xl font-bold mb-2">Error loading dashboard</h2>
        <p>{errorMessage || 'An unknown error occurred'}</p>
        <Button className="mt-4" variant="outline">
          <Link to="/dashboards">Back to Dashboards</Link>
        </Button>
      </div>
    );
  }

  if (!dashboard || !name) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <h2 className="text-xl font-bold mb-2">Dashboard not found</h2>
        <Button className="mt-4" variant="outline">
          <Link to="/dashboards">Back to Dashboards</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-8">
      <div className="flex items-center mb-8">
        <Button variant="outline" className="mr-4">
          <Link to="/dashboards" className="flex items-center">
            <ChevronLeft className="h-4 w-4 mr-2" />
            Back
          </Link>
        </Button>
        <h1 className="text-3xl font-bold">{name}</h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 mb-8">
        <div className="col-span-1 lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Display Preview</CardTitle>
              <CardDescription>How your dashboard looks on the splitflap display</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-64">
                {size ? (
                  <DisplayPreview 
                    width={size.width} 
                    height={size.height}
                    routineLocations={dashboard.routines.map((r, idx) => ({
                      id: `routine-${idx}`,
                      type: r.type,
                      location: r.location,
                      size: r.size,
                    }))}
                  />
                ) : (
                  <div className="h-full bg-muted rounded-md flex items-center justify-center">
                    <p className="text-muted-foreground">Display size not available</p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        <div>
          <Card>
            <CardHeader>
              <CardTitle>Dashboard Info</CardTitle>
              <CardDescription>Configuration details</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div>
                  <span className="font-semibold">Name:</span> {name}
                </div>
                <div>
                  <span className="font-semibold">Routines:</span> {dashboard.routines.length}
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-2">
              <ActivateDashboardButton 
                dashboardName={name} 
                fullWidth
              />
              <Button className="w-full" variant="outline">
                <Link to={`/dashboards/${name}/edit`} className="w-full">Edit Dashboard</Link>
              </Button>
              <DeleteDashboardButton 
                dashboardName={name} 
                fullWidth
                variant="ghost" 
                redirectTo="/dashboards"
              />
            </CardFooter>
          </Card>
        </div>
      </div>

      <div className="mb-8">
        <h2 className="text-2xl font-bold mb-4">Routines</h2>
        {dashboard.routines.length === 0 ? (
          <div className="bg-muted rounded-md p-8 text-center">
            <p className="text-muted-foreground mb-4">This dashboard has no routines yet.</p>
            <Button>
              <Link to={`/dashboards/${name}/edit`}>Add Routine</Link>
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {dashboard.routines.map((routine, index) => (
              <Card key={index} className="overflow-hidden">
                <CardHeader>
                  <CardTitle>{routine.type}</CardTitle>
                  <CardDescription>{RoutineTypeLabels[routine.type] || routine.type}</CardDescription>
                </CardHeader>
                <CardContent>
                  <div>
                    <div className="mb-2">
                      <span className="font-semibold">Type:</span> {routine.type}
                    </div>
                    <div className="mb-2">
                      <span className="font-semibold">Description:</span>{' '}
                      {RoutineTypeDescriptions[routine.type] || 'Custom routine'}
                    </div>
                    
                    {/* Show specific routine details based on type */}
                    {routine.type === 'TEXT' && (
                      <div className="mb-2">
                        <span className="font-semibold">Text:</span> {routine.routine.text}
                      </div>
                    )}
                    {routine.type === 'CLOCK' && (
                      <div className="mb-2">
                        <span className="font-semibold">Format:</span> {routine.routine.format}
                      </div>
                    )}
                    {routine.type === 'TIMER' && (
                      <div>
                        <div className="mb-2">
                          <span className="font-semibold">Countdown:</span> {routine.routine.countdown}s
                        </div>
                        <div className="mb-2">
                          <span className="font-semibold">Format:</span> {routine.routine.format}
                        </div>
                      </div>
                    )}
                    {routine.type === 'SLOWTEXT' && (
                      <div>
                        <div className="mb-2">
                          <span className="font-semibold">Text:</span> {routine.routine.text}
                        </div>
                        <div className="mb-2">
                          <span className="font-semibold">Delay:</span> {routine.routine.delay}ms
                        </div>
                      </div>
                    )}
                    {routine.type === 'WEATHER' && (
                      <div>
                        <div className="mb-2">
                          <span className="font-semibold">Location:</span> {routine.routine.location}
                        </div>
                      </div>
                    )}
                    
                    {/* Display location and size info */}
                    <div className="mt-4 p-3 bg-muted rounded-md">
                      <div className="text-sm text-muted-foreground">
                        Position: {routine.location.x},{routine.location.y}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        Size: {routine.size.width}×{routine.size.height}
                      </div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}