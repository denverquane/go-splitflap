import { useNavigate } from 'react-router';
import { useCreateOrUpdateDashboard } from '@/components/go-splitflap/hooks';
import DashboardBuilder from '@/components/go-splitflap/DashboardBuilder';
import { toast } from '@/components/shadcn/ui/use-toast';

export default function NewDashboardPage() {
  const navigate = useNavigate();
  const { mutateAsync: createDashboard } = useCreateOrUpdateDashboard();

  const handleSave = async (name: string, routines: any[]) => {
    try {
      console.log("Saving dashboard with routines:", routines);
      
      // Format the data for the API
      const dashboardData = routines.map(routine => {
          // Extract configs that should be sent to the backend
          // Filter out any type/internal properties that the API doesn't need
          const routineConfig = {...routine.config};
          
          // Delete any properties that start with '_' (internal)
          Object.keys(routineConfig).forEach(key => {
            if (key.startsWith('_') || key === 'type') {
              delete routineConfig[key];
            }
          });
          
          console.log(`Routine ${routine.type} config:`, routineConfig);
          
          return {
            type: routine.type,
            location: routine.location,
            size: routine.size,
            routine: routineConfig
          };
        });
    
      
      console.log("Sending to API:", JSON.stringify(dashboardData, null, 2));

      // Create the dashboard
      await createDashboard({ name, dashboard: dashboardData });

      // Show success message
      toast({
        title: 'Dashboard Created',
        description: `Dashboard "${name}" has been created successfully.`,
        variant: 'default',
      });

      // Navigate back to dashboards list
      navigate('/dashboards');
    } catch (error) {
      console.error('Error creating dashboard:', error);
      
      // Show error message
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to create dashboard',
        variant: 'destructive',
      });
    }
  };

  const handleCancel = () => {
    navigate('/dashboards');
  };

  return (
    <div className="container mx-auto py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Create New Dashboard</h1>
        <p className="text-muted-foreground mt-2">
          Design your dashboard by adding and positioning routines on the display.
        </p>
      </div>

      <DashboardBuilder 
        onSave={handleSave}
        onCancel={handleCancel}
      />
    </div>
  );
}