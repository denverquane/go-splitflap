import React, { useMemo } from 'react';
import { generateRandomColor } from '@/utils/colors';
import { Size, Location } from '@/models/Dashboard';

interface DisplayPreviewProps {
  width: number;
  height: number;
  routineLocations?: Array<{
    id?: string; // Optional unique identifier
    type: string;
    location: Location;
    size: Size;
  }>;
}

const DisplayPreview: React.FC<DisplayPreviewProps> = ({ width, height, routineLocations = [] }) => {
  // Ensure routineLocations is an array (could be undefined from parent component)
  const locations = Array.isArray(routineLocations) ? routineLocations : [];
  // Create a map to track which cells are affected by which routines
  // Use useMemo to avoid unnecessary recalculations on re-renders
  const { cellMap, routineColors } = useMemo(() => {
    const cellMap = Array(height).fill(0).map(() => Array(width).fill(null));
    const routineColors = new Map();
    
    // Fill the map with routine data
    locations.forEach((routine, routineIndex) => {
      const { x, y } = routine.location;
      const { width: routineWidth, height: routineHeight } = routine.size;
      const routineColor = generateRandomColor(routineIndex);
      
      // Use type as the identifier and create an ID if not provided
      const routineId = routine.id || `routine-${routineIndex}`;
      
      // Store the color for this routine
      routineColors.set(routineId, {
        color: routineColor,
        type: routine.type
      });
      
      // Add this routine to all cells it covers
      for (let ry = 0; ry < routineHeight; ry++) {
        for (let rx = 0; rx < routineWidth; rx++) {
          const cellY = y + ry;
          const cellX = x + rx;
          
          // Check that we're within the display bounds
          if (cellY >= 0 && cellY < height && cellX >= 0 && cellX < width) {
            cellMap[cellY][cellX] = {
              routineId: routineId,
              type: routine.type,
              color: routineColor,
            };
          }
        }
      }
    });
    
    return { cellMap, routineColors };
  }, [width, height, locations]);

  return (
    <div className="flex flex-col items-center justify-center bg-muted rounded-md p-2 h-full w-full">
      <div 
        className="grid gap-1 bg-background"
        style={{
          gridTemplateColumns: `repeat(${width}, 1fr)`,
          gridTemplateRows: `repeat(${height}, 1fr)`,
          aspectRatio: width / height > 0 ? width / height : 1,
          width: '100%',
          maxHeight: '100%'
        }}
      >
        {cellMap.flat().map((cell, index) => {
          const rowIndex = Math.floor(index / width);
          const colIndex = index % width;
          const positionLabel = `${colIndex},${rowIndex}`;
          
          if (cell) {
            // This cell is part of a routine
            return (
              <div 
                key={index}
                className="border border-border flex items-center justify-center overflow-hidden"
                style={{ 
                  aspectRatio: '1/1',
                  backgroundColor: cell.color,
                  position: 'relative',
                }}
                title={`${cell.type} at ${positionLabel}`}
              >
                <span className="text-xs text-white font-bold truncate px-1 uppercase">
                  {cell.type.charAt(0)}
                </span>
              </div>
            );
          } else {
            // Empty cell
            return (
              <div 
                key={index}
                className="bg-secondary border border-border"
                style={{ aspectRatio: '1/1' }}
                title={`Empty cell at ${positionLabel}`}
              />
            );
          }
        })}
      </div>
    </div>
  );
};

export default DisplayPreview;