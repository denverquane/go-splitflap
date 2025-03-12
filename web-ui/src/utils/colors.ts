/**
 * Generates a visually distinct color based on an index
 * Uses a predefined list of vibrant colors that work well against dark and light backgrounds
 */
export function generateRandomColor(index: number): string {
  // A list of vibrant, accessible colors that work well for visualization
  const colors = [
    '#3498db', // Blue
    '#e74c3c', // Red
    '#2ecc71', // Green
    '#f39c12', // Orange
    '#9b59b6', // Purple
    '#1abc9c', // Teal
    '#d35400', // Burnt Orange
    '#8e44ad', // Violet
    '#27ae60', // Emerald
    '#e67e22', // Carrot
    '#16a085', // Green Sea
    '#c0392b', // Pomegranate
    '#2980b9', // Belize Hole
    '#f1c40f', // Sunflower
  ];
  
  // Pick a color from the array based on the index
  return colors[index % colors.length];
}

/**
 * Generates a pastel version of a color
 */
export function generatePastelColor(index: number): string {
  const colors = [
    '#a8d8ff', // Light Blue
    '#ffb3b3', // Light Red
    '#b3ffb3', // Light Green
    '#ffda99', // Light Orange
    '#d9b3ff', // Light Purple
    '#b3fff0', // Light Teal
    '#ffb380', // Light Burnt Orange
    '#cc99ff', // Light Violet
    '#99ffcc', // Light Emerald
    '#ffcc99', // Light Carrot
    '#99ffe6', // Light Green Sea
    '#ff9999', // Light Pomegranate
    '#99ccff', // Light Belize Hole
    '#fff099', // Light Sunflower
  ];
  
  return colors[index % colors.length];
}