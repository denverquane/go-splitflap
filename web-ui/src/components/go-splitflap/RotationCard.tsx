import React from 'react';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import ActivateRotationButton from './ActivateRotationButton';
import DeleteRotationButton from './DeleteRotationButton';
import { Button } from '@/components/shadcn/ui/button';
import { Link } from 'react-router';
import { Rotation } from '@/models/Rotation';
import { Clock } from 'lucide-react';

interface RotationCardProps {
  name: string;
  rotation: Rotation;
  isActive?: boolean;
}

const RotationCard: React.FC<RotationCardProps> = ({ name, rotation, isActive = false }) => {
  // Format duration in seconds to a readable format
  const formatDuration = (seconds: number): string => {
    if (seconds < 60) {
      return `${seconds} sec`;
    } else if (seconds < 3600) {
      const minutes = Math.floor(seconds / 60);
      const remainingSeconds = seconds % 60;
      return remainingSeconds > 0 
        ? `${minutes} min ${remainingSeconds} sec` 
        : `${minutes} min`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const remainingMinutes = Math.floor((seconds % 3600) / 60);
      return remainingMinutes > 0 
        ? `${hours} hr ${remainingMinutes} min`
        : `${hours} hr`;
    }
  };

  // Handle rotation data possibly being undefined
  const entries = rotation?.rotation || [];
  
  // Debug the data coming into the component
  console.log(`Rotation card for ${name}:`, rotation);

  return (
    <Card className={`overflow-hidden ${isActive ? 'border-2 border-primary shadow-lg' : ''}`}>
      <CardHeader className={isActive ? 'bg-primary/10' : ''}>
        <div className="flex items-center justify-between">
          <CardTitle>{name}</CardTitle>
          {isActive && (
            <span className="px-2 py-1 bg-primary text-primary-foreground text-xs font-medium rounded-full">
              Active
            </span>
          )}
        </div>
        <CardDescription>Dashboard Rotation</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2 max-h-48 overflow-y-auto pr-2">
          {entries.map((entry, index) => (
            <div 
              key={`${entry.name}-${index}`} 
              className="p-2 bg-muted rounded-md flex justify-between items-center"
            >
              <div className="font-medium truncate mr-2">
                {entry.name}
              </div>
              <div className="flex items-center text-sm text-muted-foreground whitespace-nowrap">
                <Clock className="h-3 w-3 mr-1" />
                {formatDuration(entry.duration_secs)}
              </div>
            </div>
          ))}
          
          {entries.length === 0 && (
            <div className="p-4 bg-muted rounded-md text-center text-muted-foreground">
              No dashboards in this rotation
            </div>
          )}
        </div>
      </CardContent>
      <CardFooter className="flex flex-col gap-2">
        <ActivateRotationButton 
          rotationName={name} 
          fullWidth
          isActive={isActive}
        />
        <div className="flex justify-between w-full gap-2">
          <Button variant="outline" className="flex-1">
            <Link to={`/rotations/${name}/edit`} className="w-full">Edit</Link>
          </Button>
          <Button variant="secondary" className="flex-1">
            <Link to={`/rotations/${name}`} className="w-full">View</Link>
          </Button>
        </div>
        <DeleteRotationButton 
          rotationName={name} 
          fullWidth
          variant="ghost" 
          size="sm"
        />
      </CardFooter>
    </Card>
  );
};

export default RotationCard;