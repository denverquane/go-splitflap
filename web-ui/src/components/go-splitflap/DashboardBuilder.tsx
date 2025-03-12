import React, { useState, useRef, useEffect } from 'react';
import { useRoutines } from './hooks/use-routines';
import { useDisplaySize } from './hooks/use-dashboards';
import { generateRandomColor } from '@/utils/colors';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { Button } from '@/components/shadcn/ui/button';
import { Input } from '@/components/shadcn/ui/input';
import { Label } from '@/components/shadcn/ui/label';
import { 
  Select, 
  SelectContent, 
  SelectItem, 
  SelectTrigger, 
  SelectValue 
} from '@/components/shadcn/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/shadcn/ui/tabs';
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
} from "@/components/shadcn/ui/alert-dialog";
import { Plus, Trash2, MoveVertical, MoveHorizontal, Save, HelpCircle, Settings, Layers } from 'lucide-react';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/shadcn/ui/tooltip';
import RoutineParamsEditor from './RoutineParamsEditor';

// Types for our component
interface Location {
  x: number;
  y: number;
}

interface Size {
  width: number;
  height: number;
}

interface DashboardRoutine {
  id: string; // Unique identifier for this instance
  type: string; // The routine type identifier used by backend
  location: Location;
  size: Size;
  config: Record<string, any>; // Configuration parameters for the routine
  routine?: Record<string, any>; // Used when sending to API
}

interface DashboardBuilderProps {
  initialName?: string;
  initialRoutines?: DashboardRoutine[];
  onSave: (name: string, routines: DashboardRoutine[]) => void;
  onCancel?: () => void;
}

const DashboardBuilder: React.FC<DashboardBuilderProps> = ({ 
  initialName = '',
  initialRoutines = [], 
  onSave,
  onCancel
}) => {
  // State for dashboard
  const [dashboardName, setDashboardName] = useState(initialName);
  const [routines, setRoutines] = useState<DashboardRoutine[]>(initialRoutines);

  // Selected routine for editing
  const [selectedRoutineId, setSelectedRoutineId] = useState<string | null>(null);
  const [draggedRoutineId, setDraggedRoutineId] = useState<string | null>(null);
  const [resizingRoutineId, setResizingRoutineId] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<string>("dashboard");

  // Grid parameters
  const [gridDimensions, setGridDimensions] = useState({ width: 0, height: 0 });
  const [cellSize, setCellSize] = useState(30); // Default cell size
  const gridRef = useRef<HTMLDivElement>(null);

  // Get available routines from API
  const { routines: availableRoutines, routinesData, isLoading: routinesLoading } = useRoutines();
  const { size: displaySize, isLoading: sizeLoading } = useDisplaySize();

  // Start position for drag operations
  const dragStart = useRef<{ x: number, y: number }>({ x: 0, y: 0 });
  const resizeStart = useRef<{ width: number, height: number }>({ width: 0, height: 0 });

  // Colors for routines
  const routineColors = useRef(new Map<string, string>());

  // Setup grid dimensions when display size is loaded
  useEffect(() => {
    if (displaySize) {
      setGridDimensions({
        width: displaySize.width,
        height: displaySize.height
      });
    }
  }, [displaySize]);

  // Calculate cell size based on grid dimensions and container size
  useEffect(() => {
    if (gridRef.current && gridDimensions.width && gridDimensions.height) {
      const containerWidth = gridRef.current.offsetWidth - 32; // Account for padding
      const containerHeight = gridRef.current.offsetHeight - 32;
      
      // Calculate cell size to fit the container
      const cellWidth = Math.floor(containerWidth / gridDimensions.width);
      const cellHeight = Math.floor(containerHeight / gridDimensions.height);
      
      // Use the smaller of the two to ensure grid fits
      const newCellSize = Math.max(10, Math.min(cellWidth, cellHeight));
      
      // Only update if cell size changes to avoid re-renders
      if (newCellSize !== cellSize) {
        setCellSize(newCellSize);
      }
    }
  }, [gridDimensions, gridRef.current?.offsetWidth, gridRef.current?.offsetHeight]);
  
  // Add resize observer to recalculate when container size changes
  useEffect(() => {
    if (!gridRef.current) return;
    
    const resizeObserver = new ResizeObserver(() => {
      if (gridRef.current && gridDimensions.width && gridDimensions.height) {
        const containerWidth = gridRef.current.offsetWidth - 32;
        const containerHeight = gridRef.current.offsetHeight - 32;
        
        const cellWidth = Math.floor(containerWidth / gridDimensions.width);
        const cellHeight = Math.floor(containerHeight / gridDimensions.height);
        
        const newCellSize = Math.max(10, Math.min(cellWidth, cellHeight));
        if (newCellSize !== cellSize) {
          setCellSize(newCellSize);
        }
      }
    });
    
    resizeObserver.observe(gridRef.current);
    
    return () => {
      resizeObserver.disconnect();
    };
  }, [gridDimensions.width, gridDimensions.height, cellSize]);

  // Generate consistent colors for routines
  useEffect(() => {
    routines.forEach((routine, index) => {
      if (!routineColors.current.has(routine.id)) {
        routineColors.current.set(routine.id, generateRandomColor(index));
      }
    });
  }, [routines]);

  // Add a new routine to the dashboard
  const addRoutine = () => {
    if (availableRoutines.length === 0) return;
    
    const routineType = availableRoutines[0];
    const routineData = routinesData[routineType];
    
    // Get size constraints for this routine
    const minWidth = routineData.min_size?.width || 1;
    const minHeight = routineData.min_size?.height || 1;
    
    // Use min_size as default or fallback to 3x1
    const initialWidth = minWidth || 3;
    const initialHeight = minHeight || 1;
    
    // Initialize config with default values for parameters
    const initialConfig: Record<string, any> = { ...routineData.config };
    
    // If parameters exist, set initial values based on type
    if (routineData.parameters && routineData.parameters.length > 0) {
      routineData.parameters.forEach(param => {
        // Set default values based on parameter type
        switch (param.type.toLowerCase()) {
          case 'bool':
            initialConfig[param.field] = false;
            break;
          case 'int':
          case 'int64':
          case 'float':
          case 'float64':
            initialConfig[param.field] = 0;
            break;
          case 'string[]':
            initialConfig[param.field] = [];
            break;
          case 'string':
            // Check if there are options in the description
            const options = param.description.match(/options:\s*([^.]+)/i);
            if (options && options[1]) {
              // Use the first option as default
              const firstOption = options[1].split(',')[0].trim();
              initialConfig[param.field] = firstOption;
            } else {
              initialConfig[param.field] = '';
            }
            break;
          default:
            initialConfig[param.field] = '';
        }
      });
    }
    
    const newRoutine: DashboardRoutine = {
      id: `routine-${Date.now()}-${Math.random().toString(36).substring(2, 9)}`,
      type: routineType,
      location: { x: 0, y: 0 },
      size: { width: initialWidth, height: initialHeight },
      config: initialConfig,
    };
    
    setRoutines([...routines, newRoutine]);
    setSelectedRoutineId(newRoutine.id);
  };

  // Remove a routine from the dashboard
  const removeRoutine = (id: string) => {
    setRoutines(routines.filter(r => r.id !== id));
    if (selectedRoutineId === id) {
      setSelectedRoutineId(null);
    }
  };

  // Select a routine for editing
  const selectRoutine = (id: string) => {
    setSelectedRoutineId(id);
    setActiveTab("routines");
  };

  // Handle routine property changes
  const updateRoutineProperty = (id: string, property: string, value: any) => {
    setRoutines(routines.map(routine => 
      routine.id === id 
        ? { ...routine, [property]: value } 
        : routine
    ));
  };

  // Handle start of dragging a routine
  const handleDragStart = (e: React.MouseEvent, id: string) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Find the routine being dragged
    const routine = routines.find(r => r.id === id);
    if (!routine) return;
    
    // Select the routine and mark it as being dragged
    setSelectedRoutineId(id);
    setDraggedRoutineId(id);
    
    // Store the initial mouse position for calculating deltas
    const initialX = e.clientX;
    const initialY = e.clientY;
    
    // Store the initial routine location
    const initialLocation = { ...routine.location };
    
    // Define the drag move handler
    const dragMoveHandler = (moveEvent: MouseEvent) => {
      // Calculate the distance moved in pixels
      const deltaXPx = moveEvent.clientX - initialX;
      const deltaYPx = moveEvent.clientY - initialY;
      
      // Convert to grid cells (round to nearest cell)
      const deltaXCells = Math.round(deltaXPx / cellSize);
      const deltaYCells = Math.round(deltaYPx / cellSize);
      
      // Check if movement is significant enough to update
      if (deltaXCells === 0 && deltaYCells === 0) return;
      
      // Calculate new position with bounds checking
      const newX = Math.max(0, Math.min(
        gridDimensions.width - routine.size.width,
        initialLocation.x + deltaXCells
      ));
      
      const newY = Math.max(0, Math.min(
        gridDimensions.height - routine.size.height,
        initialLocation.y + deltaYCells
      ));
      
      // Update the routines array with the new location
      setRoutines(currentRoutines => 
        currentRoutines.map(r => 
          r.id === id 
            ? { ...r, location: { x: newX, y: newY } }
            : r
        )
      );
    };
    
    // Define the drag end handler
    const dragEndHandler = () => {
      // Clean up by removing event listeners
      document.removeEventListener('mousemove', dragMoveHandler);
      document.removeEventListener('mouseup', dragEndHandler);
      
      // Reset drag state
      setDraggedRoutineId(null);
    };
    
    // Add event listeners for move and end events
    document.addEventListener('mousemove', dragMoveHandler);
    document.addEventListener('mouseup', dragEndHandler);
  };
  
  // The handleDragMove is now defined inline in handleDragStart
  const handleDragMove = () => {}; // Empty placeholder, not used

  // This handler is now unused since we define it inline
  // But we'll keep it as an empty function to avoid changing too much code
  const handleDragEnd = () => {};

  // Handle start of resizing a routine
  const handleResizeStart = (e: React.MouseEvent, id: string) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Find the routine being resized
    const routine = routines.find(r => r.id === id);
    if (!routine) return;
    
    // Select the routine and mark it as being resized
    setSelectedRoutineId(id);
    setResizingRoutineId(id);
    
    // Store the initial mouse position
    const initialX = e.clientX;
    const initialY = e.clientY;
    
    // Store the initial size
    const initialSize = { ...routine.size };
    
    // Get the routine constraints
    const routineData = routinesData[routine.type];
    const minWidth = routineData?.min_size?.width || 1;
    const minHeight = routineData?.min_size?.height || 1;
    const maxWidth = Math.min(
      routineData?.max_size?.width || Number.MAX_SAFE_INTEGER,
      gridDimensions.width - routine.location.x
    );
    const maxHeight = Math.min(
      routineData?.max_size?.height || Number.MAX_SAFE_INTEGER,
      gridDimensions.height - routine.location.y
    );
    
    // Define the resize move handler
    const resizeMoveHandler = (moveEvent: MouseEvent) => {
      // Calculate the distance moved in pixels
      const deltaXPx = moveEvent.clientX - initialX;
      const deltaYPx = moveEvent.clientY - initialY;
      
      // Convert to grid cells (round to nearest cell)
      const deltaWidthCells = Math.round(deltaXPx / cellSize);
      const deltaHeightCells = Math.round(deltaYPx / cellSize);
      
      // Check if significant movement
      if (deltaWidthCells === 0 && deltaHeightCells === 0) return;
      
      // Calculate new size with constraints
      const newWidth = Math.max(minWidth, Math.min(
        maxWidth,
        initialSize.width + deltaWidthCells
      ));
      
      const newHeight = Math.max(minHeight, Math.min(
        maxHeight,
        initialSize.height + deltaHeightCells
      ));
      
      // Update the routines array with the new size
      setRoutines(currentRoutines => 
        currentRoutines.map(r => 
          r.id === id 
            ? { ...r, size: { width: newWidth, height: newHeight } }
            : r
        )
      );
    };
    
    // Define the resize end handler
    const resizeEndHandler = () => {
      // Clean up
      document.removeEventListener('mousemove', resizeMoveHandler);
      document.removeEventListener('mouseup', resizeEndHandler);
      document.body.classList.remove('resizing');
      
      // Reset resize state
      setResizingRoutineId(null);
    };
    
    // Add event listeners
    document.addEventListener('mousemove', resizeMoveHandler);
    document.addEventListener('mouseup', resizeEndHandler);
    document.body.classList.add('resizing');
  };
  
  // The handleResizeMove is now defined inline in handleResizeStart
  const handleResizeMove = () => {}; // Empty placeholder, not used

  // This handler is now unused since we define it inline
  // But we'll keep it as an empty function to avoid changing too much code
  const handleResizeEnd = () => {};

  // Save the dashboard
  const handleSave = () => {
    if (!dashboardName.trim()) {
      alert("Please provide a dashboard name");
      return;
    }
    
    // Pass the raw routines to the parent component
    console.log("Preparing routines for saving:", routines);
    onSave(dashboardName, routines);
  };

  // Check if routines are loaded
  if (routinesLoading || sizeLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary"></div>
      </div>
    );
  }

  // Check if there are any available routines
  if (availableRoutines.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64">
        <p className="text-lg font-medium mb-4">No routines available</p>
        <p className="text-muted-foreground mb-6">
          Create some routines first before creating a dashboard.
        </p>
        <Button>
          <a href="/routines">Go to Routines</a>
        </Button>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-4">
      <div className="space-y-6">
        {/* Dashboard Name */}
        <Card>
          <CardHeader className="p-4">
            <CardTitle>Dashboard Settings</CardTitle>
          </CardHeader>
          <CardContent className="p-4 pt-0">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="dashboard-name">Dashboard Name</Label>
                <Input 
                  id="dashboard-name" 
                  value={dashboardName} 
                  onChange={(e) => setDashboardName(e.target.value)}
                  placeholder="Enter dashboard name" 
                />
              </div>
              
              <div className="space-y-2">
                <Label>Display Size</Label>
                <div className="flex items-center space-x-2 text-sm">
                  <span className="bg-muted px-2 py-1 rounded">
                    Width: {gridDimensions.width} cells
                  </span>
                  <span className="bg-muted px-2 py-1 rounded">
                    Height: {gridDimensions.height} cells
                  </span>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Preview and Routines in two-column layout */}
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-6">
          {/* Preview - Takes 3/5 of the space on large screens */}
          <Card className="lg:col-span-3">
            <CardHeader className="p-4">
              <CardTitle className="text-base">Preview</CardTitle>
            </CardHeader>
            <CardContent className="p-4 pt-0" style={{ height: '500px' }}>
              <div 
                ref={gridRef} 
                className="bg-muted rounded-md p-4 w-full h-full flex items-center justify-center overflow-hidden"
              >
                <div
                  className="bg-background relative overflow-hidden"
                  style={{
                    display: 'grid',
                    gridTemplateColumns: `repeat(${gridDimensions.width}, ${cellSize}px)`,
                    gridTemplateRows: `repeat(${gridDimensions.height}, ${cellSize}px)`,
                    gap: '1px',
                    maxWidth: '100%',
                    maxHeight: '100%',
                    width: 'min-content',
                    height: 'min-content',
                  }}
                >
                  {/* Background grid */}
                  {Array.from({ length: gridDimensions.width * gridDimensions.height }).map((_, index) => {
                    const x = index % gridDimensions.width;
                    const y = Math.floor(index / gridDimensions.width);
                    
                    return (
                      <div
                        key={`cell-${x}-${y}`}
                        className="bg-secondary border border-border"
                        style={{
                          width: `${cellSize}px`,
                          height: `${cellSize}px`,
                        }}
                      />
                    );
                  })}
                  
                  {/* Routine elements */}
                  {routines.map((routine) => {
                    const color = routineColors.current.get(routine.id) || '#ccc';
                    const routineData = routinesData[routine.type];

                    // Check if at size limits
                    const minWidth = routineData?.min_size?.width || 1;
                    const minHeight = routineData?.min_size?.height || 1;
                    const maxWidth = Math.min(
                      routineData?.max_size?.width || Number.MAX_SAFE_INTEGER,
                      gridDimensions.width - routine.location.x
                    );
                    const maxHeight = Math.min(
                      routineData?.max_size?.height || Number.MAX_SAFE_INTEGER,
                      gridDimensions.height - routine.location.y
                    );

                    const isAtMinWidth = routine.size.width <= minWidth;
                    const isAtMinHeight = routine.size.height <= minHeight;
                    const isAtMaxWidth = routine.size.width >= maxWidth;
                    const isAtMaxHeight = routine.size.height >= maxHeight;
                    const isAtSizeLimit = isAtMinWidth || isAtMinHeight || isAtMaxWidth || isAtMaxHeight;
                    
                    return (
                      <div
                        key={routine.id}
                        className={`absolute flex items-center justify-center cursor-move border-2 transition-colors shadow-sm
                          ${draggedRoutineId === routine.id ? 'opacity-70 shadow-lg' : 'opacity-100'}
                          ${resizingRoutineId === routine.id ? 'opacity-80 shadow-lg' : ''}
                          ${selectedRoutineId === routine.id ? 'border-primary z-10' : 'border-transparent z-1'}
                          ${isAtSizeLimit && selectedRoutineId === routine.id ? 'ring-2 ring-yellow-500/50' : ''}
                          select-none touch-none
                        `}
                        style={{
                          backgroundColor: color,
                          left: `${routine.location.x * cellSize}px`,
                          top: `${routine.location.y * cellSize}px`,
                          width: `${routine.size.width * cellSize}px`,
                          height: `${routine.size.height * cellSize}px`,
                        }}
                        onClick={() => selectRoutine(routine.id)}
                        onMouseDown={(e) => handleDragStart(e, routine.id)}
                        title={isAtSizeLimit ? `This routine has reached a size limit (min: ${minWidth}x${minHeight}, max: ${maxWidth}x${maxHeight})` : undefined}
                      >
                        <div className="text-xs font-medium text-white truncate px-2">
                          {routine.type}
                        </div>
                        
                        {/* Size constraint indicators */}
                        {selectedRoutineId === routine.id && (
                          <>
                            {isAtMinWidth && (
                              <div className="absolute left-0 top-0 bottom-0 w-1 bg-yellow-500/50" 
                                title="At minimum width limit"/>
                            )}
                            {isAtMinHeight && (
                              <div className="absolute top-0 left-0 right-0 h-1 bg-yellow-500/50"
                                title="At minimum height limit"/>
                            )}
                            {isAtMaxWidth && (
                              <div className="absolute right-0 top-0 bottom-0 w-1 bg-yellow-500/50"
                                title="At maximum width limit"/>
                            )}
                            {isAtMaxHeight && (
                              <div className="absolute bottom-0 left-0 right-0 h-1 bg-yellow-500/50"
                                title="At maximum height limit"/>
                            )}
                          </>
                        )}
                        
                        {/* Resize handle */}
                        <div
                          className={`absolute bottom-0 right-0 w-6 h-6 cursor-se-resize bg-white/30 flex items-center justify-center
                            hover:bg-white/50 active:bg-white/70 rounded-tl touch-none select-none
                            ${isAtMaxWidth || isAtMaxHeight ? 'opacity-50' : ''}
                          `}
                          role="button"
                          aria-label="Resize routine"
                          tabIndex={0}
                          style={{ touchAction: 'none', userSelect: 'none' }}
                          onMouseDown={(e) => handleResizeStart(e, routine.id)}
                          onClick={(e) => e.stopPropagation()}
                          onTouchStart={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            // Create a synthetic mousedown event
                            const touch = e.touches[0];
                            const mouseEvent = new MouseEvent('mousedown', {
                              clientX: touch.clientX,
                              clientY: touch.clientY,
                              bubbles: true,
                              cancelable: true,
                              view: window
                            });
                            e.currentTarget.dispatchEvent(mouseEvent);
                          }}
                        >
                          <svg width="12" height="12" viewBox="0 0 8 8" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M7 1L1 7M7 4L4 7M7 7L7 7" stroke="white" strokeWidth="2"/>
                          </svg>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Routines Panel - Takes 2/5 of the space on large screens */}
          <div className="lg:col-span-2 space-y-4">
            <div className="flex justify-between items-center">
              <h2 className="text-lg font-medium">Dashboard Routines</h2>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button onClick={addRoutine} size="sm">
                      <Plus className="h-4 w-4 mr-1" />
                      Add Routine
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    Add a new routine to this dashboard
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            
            {/* Routines List */}
            <Card className="h-[220px] overflow-auto">
              <CardHeader className="p-3">
                <CardTitle className="text-sm">Routines</CardTitle>
              </CardHeader>
              <CardContent className="p-3 pt-0">
                {routines.length === 0 ? (
                  <div className="flex flex-col items-center justify-center h-32 text-muted-foreground">
                    <p>No routines added yet</p>
                    <Button onClick={addRoutine} variant="link" className="mt-2">
                      Add your first routine
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {routines.map((routine) => (
                      <div 
                        key={routine.id}
                        className={`border rounded-md p-2 cursor-pointer flex items-center justify-between
                          ${selectedRoutineId === routine.id ? 'border-primary bg-primary/5' : 'border-border'}
                        `}
                        onClick={() => selectRoutine(routine.id)}
                      >
                        <div className="flex items-center">
                          <div 
                            className="w-3 h-3 rounded-sm mr-2" 
                            style={{ backgroundColor: routineColors.current.get(routine.id) || '#ccc' }}
                          />
                          <div>
                            <div className="font-medium text-sm">{routine.type}</div>
                            <div className="text-xs text-muted-foreground">
                              {routine.size.width}×{routine.size.height} at ({routine.location.x},{routine.location.y})
                            </div>
                          </div>
                        </div>
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-6 w-6 text-muted-foreground hover:text-destructive"
                              onClick={(e) => e.stopPropagation()}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>Remove routine?</AlertDialogTitle>
                              <AlertDialogDescription>
                                Are you sure you want to remove "{routine.type}" from this dashboard?
                                This action cannot be undone.
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>Cancel</AlertDialogCancel>
                              <AlertDialogAction 
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                                onClick={() => removeRoutine(routine.id)}
                              >
                                Remove
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
            
            {/* Selected Routine Settings */}
            <Card className="h-[230px] overflow-auto">
              <CardHeader className="p-3">
                <CardTitle className="text-sm">Routine Settings</CardTitle>
              </CardHeader>
              <CardContent className="p-3 pt-0">
                {selectedRoutineId ? (
                  (() => {
                    const selectedRoutine = routines.find(r => r.id === selectedRoutineId);
                    if (!selectedRoutine) return null;
                    
                    // Get the routine data
                    const routineData = routinesData[selectedRoutine.type];
                    
                    return (
                      <div className="space-y-3">
                        {/* Routine Type Selector */}
                        <div className="space-y-2">
                          <Label htmlFor="routine-type" className="text-xs">Routine Type</Label>
                          <Select
                            value={selectedRoutine.type}
                            onValueChange={(value) => {
                              // Skip if selecting the same routine type
                              if (value === selectedRoutine.type) return;
                              
                              console.log("Changing routine type to:", value);
                              const routineData = routinesData[value];
                              
                              // Get min size for the new routine type
                              const minWidth = routineData.min_size?.width || 1;
                              const minHeight = routineData.min_size?.height || 1;
                              
                              // Make sure current size meets minimum requirements
                              const newWidth = Math.max(minWidth, selectedRoutine.size.width);
                              const newHeight = Math.max(minHeight, selectedRoutine.size.height);
                              
                              // Initialize config with default values for parameters
                              const initialConfig: Record<string, any> = { ...routineData.config };
                              
                              // If parameters exist, set initial values based on type
                              if (routineData.parameters && routineData.parameters.length > 0) {
                                routineData.parameters.forEach(param => {
                                  // Set default values based on parameter type
                                  switch (param.type.toLowerCase()) {
                                    case 'bool':
                                      initialConfig[param.field] = false;
                                      break;
                                    case 'int':
                                    case 'int64':
                                    case 'float':
                                    case 'float64':
                                      initialConfig[param.field] = 0;
                                      break;
                                    case 'string[]':
                                      initialConfig[param.field] = [];
                                      break;
                                    case 'string':
                                      // Check if there are options in the description
                                      const options = param.description.match(/options:\s*([^.]+)/i);
                                      if (options && options[1]) {
                                        // Use the first option as default
                                        const firstOption = options[1].split(',')[0].trim();
                                        initialConfig[param.field] = firstOption;
                                      } else {
                                        initialConfig[param.field] = '';
                                      }
                                      break;
                                    default:
                                      initialConfig[param.field] = '';
                                  }
                                });
                              }
                              
                              // Create a new routine object with the updated properties
                              const updatedRoutine = {
                                ...selectedRoutine,
                                type: value,
                                config: initialConfig,
                                size: {
                                  width: newWidth,
                                  height: newHeight
                                }
                              };
                              
                              // Update the routines array with this new routine
                              setRoutines(current => 
                                current.map(r => 
                                  r.id === selectedRoutineId ? updatedRoutine : r
                                )
                              );
                            }}
                          >
                            <SelectTrigger className="h-8 text-xs">
                              <SelectValue placeholder="Select a routine" />
                            </SelectTrigger>
                            <SelectContent>
                              {availableRoutines.map((name) => (
                                <SelectItem key={name} value={name}>{name}</SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </div>
                        
                        {/* Position & Size Controls */}
                        <div className="grid grid-cols-2 gap-3">
                          <div className="space-y-1">
                            <Label className="text-xs">Position</Label>
                            <div className="flex items-center space-x-1">
                              <MoveHorizontal className="h-3 w-3 text-muted-foreground" />
                              <Input
                                type="number"
                                min={0}
                                max={gridDimensions.width - selectedRoutine.size.width}
                                value={selectedRoutine.location.x}
                                onChange={(e) => {
                                  const value = Math.max(0, Math.min(
                                    gridDimensions.width - selectedRoutine.size.width,
                                    parseInt(e.target.value) || 0
                                  ));
                                  updateRoutineProperty(selectedRoutineId, 'location', { 
                                    ...selectedRoutine.location, 
                                    x: value 
                                  });
                                }}
                                className="h-7 text-xs"
                              />
                            </div>
                            <div className="flex items-center space-x-1">
                              <MoveVertical className="h-3 w-3 text-muted-foreground" />
                              <Input
                                type="number"
                                min={0}
                                max={gridDimensions.height - selectedRoutine.size.height}
                                value={selectedRoutine.location.y}
                                onChange={(e) => {
                                  const value = Math.max(0, Math.min(
                                    gridDimensions.height - selectedRoutine.size.height,
                                    parseInt(e.target.value) || 0
                                  ));
                                  updateRoutineProperty(selectedRoutineId, 'location', { 
                                    ...selectedRoutine.location, 
                                    y: value 
                                  });
                                }}
                                className="h-7 text-xs"
                              />
                            </div>
                          </div>
                          
                          <div className="space-y-1">
                            <Label className="text-xs">Size</Label>
                            {(() => {
                              // Get routine constraints
                              const routineData = routinesData[selectedRoutine.type];
                              const minWidth = routineData?.min_size?.width || 1;
                              const minHeight = routineData?.min_size?.height || 1;
                              const maxWidth = Math.min(
                                routineData?.max_size?.width || Number.MAX_SAFE_INTEGER,
                                gridDimensions.width - selectedRoutine.location.x
                              );
                              const maxHeight = Math.min(
                                routineData?.max_size?.height || Number.MAX_SAFE_INTEGER,
                                gridDimensions.height - selectedRoutine.location.y
                              );
                              
                              return (
                                <>
                                  <div className="flex items-center space-x-1">
                                    <span className="text-xs text-muted-foreground">W</span>
                                    <Input
                                      type="number"
                                      min={minWidth}
                                      max={maxWidth}
                                      value={selectedRoutine.size.width}
                                      onChange={(e) => {
                                        const value = Math.max(minWidth, Math.min(
                                          maxWidth,
                                          parseInt(e.target.value) || minWidth
                                        ));
                                        updateRoutineProperty(selectedRoutineId, 'size', { 
                                          ...selectedRoutine.size, 
                                          width: value 
                                        });
                                      }}
                                      className="h-7 text-xs"
                                    />
                                    <span className="text-[10px] text-muted-foreground">{minWidth}-{maxWidth}</span>
                                  </div>
                                  
                                  <div className="flex items-center space-x-1">
                                    <span className="text-xs text-muted-foreground">H</span>
                                    <Input
                                      type="number"
                                      min={minHeight}
                                      max={maxHeight}
                                      value={selectedRoutine.size.height}
                                      onChange={(e) => {
                                        const value = Math.max(minHeight, Math.min(
                                          maxHeight,
                                          parseInt(e.target.value) || minHeight
                                        ));
                                        updateRoutineProperty(selectedRoutineId, 'size', { 
                                          ...selectedRoutine.size, 
                                          height: value 
                                        });
                                      }}
                                      className="h-7 text-xs"
                                    />
                                    <span className="text-[10px] text-muted-foreground">{minHeight}-{maxHeight}</span>
                                  </div>
                                </>
                              );
                            })()}
                          </div>
                        </div>
                        
                        {/* Parameters Section */}
                        <div className="pt-1">
                          <div className="flex items-center justify-between mb-2">
                            <Label className="text-xs flex items-center">
                              <Settings className="h-3 w-3 mr-1" />
                              Routine Parameters
                            </Label>
                          </div>
                          
                          <div className="max-h-[120px] overflow-y-auto rounded-md border border-border p-2 bg-muted/30">
                            {routineData?.parameters && routineData.parameters.length > 0 ? (
                              <div key={`params-${selectedRoutine.id}-${selectedRoutine.type}`}>
                                {/* Debug info - remove after fixing */}
                                <div className="text-[10px] text-muted-foreground mb-2">
                                  Parameters: {JSON.stringify(routineData.parameters.map(p => p.name))}
                                </div>
                                <RoutineParamsEditor 
                                  parameters={routineData.parameters} 
                                  config={{...selectedRoutine.config}} // Pass a new object to prevent reference issues
                                  onChange={(newConfig) => {
                                    // Deep compare to avoid unnecessary updates
                                    if (JSON.stringify(newConfig) !== JSON.stringify(selectedRoutine.config)) {
                                      updateRoutineProperty(selectedRoutineId, 'config', {...newConfig});
                                    }
                                  }}
                                />
                              </div>
                            ) : (
                              <div className="text-xs text-muted-foreground py-2">
                                This routine type has no configurable parameters.
                              </div>
                            )}
                          </div>
                        </div>
                        
                        <div className="text-xs text-muted-foreground mt-1">
                          <p className="flex items-center">
                            <HelpCircle className="h-3 w-3 mr-1" />
                            Drag and resize routines directly on the preview
                          </p>
                        </div>
                      </div>
                    );
                  })()
                ) : (
                  <div className="flex flex-col items-center justify-center h-24 text-muted-foreground">
                    <p className="text-sm">No routine selected</p>
                    <p className="text-xs mt-1">Select a routine to edit its properties</p>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </div>
      </div>
      
      <div className="flex justify-end mt-6 space-x-4">
        {onCancel && (
          <Button variant="outline" onClick={onCancel}>Cancel</Button>
        )}
        <Button onClick={handleSave} disabled={!dashboardName.trim() || routines.length === 0}>
          <Save className="h-4 w-4 mr-2" />
          Save Dashboard
        </Button>
      </div>
    </div>
  );
};

export default DashboardBuilder;