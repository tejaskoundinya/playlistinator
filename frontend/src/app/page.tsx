'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { toast } from 'sonner';
import { Loader2, Music, RefreshCw } from 'lucide-react';
import { generatePlaylist } from '@/services/api';

export default function Home() {
  const [isGenerating, setIsGenerating] = useState(false);
  const [progress, setProgress] = useState(0);

  const handleGeneratePlaylist = async () => {
    setIsGenerating(true);
    try {
      const response = await generatePlaylist();
      
      if (response.success) {
        toast.success(response.message);
      } else {
        toast.error(response.message);
      }
    } catch (error) {
      toast.error('Failed to generate playlist. Please try again.');
      console.error(error);
    } finally {
      setIsGenerating(false);
      setProgress(100);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-b from-gray-900 to-gray-800 p-8">
      <div className="container mx-auto max-w-4xl">
        <Card className="bg-gray-800/50 border-gray-700">
          <CardHeader>
            <CardTitle className="text-3xl font-bold text-white flex items-center gap-2">
              <Music className="h-8 w-8" />
              Playlist Generator
            </CardTitle>
            <CardDescription className="text-gray-400">
              Generate your personalized Spotify playlist based on your Last.fm listening history
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center gap-4">
                <div className="flex-1">
                  <Progress value={progress} className="h-2" />
                </div>
                <span className="text-sm text-gray-400">{progress}%</span>
              </div>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Card className="bg-gray-700/50 border-gray-600">
                  <CardHeader>
                    <CardTitle className="text-lg text-white">Last.fm Stats</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-gray-400">Your listening history from the last 30 days</p>
                  </CardContent>
                </Card>
                <Card className="bg-gray-700/50 border-gray-600">
                  <CardHeader>
                    <CardTitle className="text-lg text-white">Spotify Playlist</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-gray-400">TK - Hot 100</p>
                  </CardContent>
                </Card>
              </div>
            </div>
          </CardContent>
          <CardFooter>
            <Button
              onClick={handleGeneratePlaylist}
              disabled={isGenerating}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white"
            >
              {isGenerating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Generating Playlist...
                </>
              ) : (
                <>
                  <Music className="mr-2 h-4 w-4" />
                  Generate Playlist
                </>
              )}
            </Button>
          </CardFooter>
        </Card>
      </div>
    </main>
  );
}
