import { Outlet } from "react-router";
import { Link } from "react-router";
import { useLocation } from "react-router";
import { ModeToggle } from "@/components/shadcn/ui/mode-toggle";
import { Toaster } from "@/components/shadcn/ui/toaster";
import DisplayStatus from "@/components/go-splitflap/DisplayStatus";

export default function AppLayout() {
    const location = useLocation();
    const showStatus = location.pathname.startsWith('/dashboards') ||
        location.pathname.startsWith('/rotations') ||
        location.pathname.startsWith('/routines');

    return (
        <div className="flex flex-col min-h-screen">
            <header className="border-b px-4 py-2">
                <div className="container mx-auto flex justify-between items-center">
                    <Link to="/" className="text-xl font-bold">SplitFlap</Link>
                    <ModeToggle />
                </div>
            </header>

            {showStatus && (
                <div className="container mx-auto pt-6 pb-0">
                    <DisplayStatus />
                </div>
            )}

            {showStatus && (
                <div className="container mx-auto py-4 border-b">
                    <nav className="flex items-center space-x-6 justify-center">
                        <div className="flex space-x-6">
                            <Link to="/routines" className="hover:text-primary font-medium">Routines</Link>
                            <Link to="/dashboards" className="hover:text-primary font-medium">Dashboards</Link>
                            <Link to="/rotations" className="hover:text-primary font-medium">Rotations</Link>
                        </div>
                    </nav>
                </div>
            )}

            <main className="flex-1">
                <Outlet />
            </main>
            <Toaster />
        </div>
    );
}