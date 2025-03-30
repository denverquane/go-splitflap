import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/shadcn/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/shadcn/ui/tabs';
import TranslationSettings from '@/components/go-splitflap/TranslationSettings';

const SettingsPage: React.FC = () => {
  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Configure your splitflap display settings</p>
      </div>

      <Tabs defaultValue="translations">
        <TabsList>
          <TabsTrigger value="translations">Character Translations</TabsTrigger>
          <TabsTrigger value="display">Display</TabsTrigger>
        </TabsList>
        
        <TabsContent value="translations" className="mt-4">
          <TranslationSettings />
        </TabsContent>
        
        <TabsContent value="display" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Display Settings</CardTitle>
              <CardDescription>
                Configure display-related settings
              </CardDescription>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">Display settings will be added in a future update.</p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
};

export default SettingsPage;