'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Music, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { generatePlaylist } from '@/services/api';

export default function Home() {
  const [isLoading, setIsLoading] = useState(false);

  const handleGeneratePlaylist = async () => {
    setIsLoading(true);
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
      setIsLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-b from-gray-900 to-black text-white">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-3xl mx-auto text-center">
          <h1 className="text-5xl font-bold mb-6 bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-purple-500">
            Playlistinator
          </h1>
          <p className="text-xl text-gray-300 mb-12">
            Transform your Last.fm listening history into a beautiful Spotify playlist
          </p>
          
          <div className="bg-gray-800/50 backdrop-blur-sm rounded-2xl p-8 shadow-xl border border-gray-700/50">
            <div className="space-y-6">
              <div className="flex items-center justify-center space-x-12 text-gray-300">
                <div className="flex flex-col items-center">
                  <div className="w-12 h-12 rounded-full bg-blue-500/20 flex items-center justify-center mb-2">
                    <Music className="w-6 h-6 text-blue-400" />
                  </div>
                  <span className="text-sm">Last.fm</span>
                </div>
                <div className="w-24 h-0.5 bg-gray-600"></div>
                <div className="flex flex-col items-center">
                  <div className="w-12 h-12 rounded-full bg-green-500/20 flex items-center justify-center mb-2">
                    <Music className="w-6 h-6 text-green-400" />
                  </div>
                  <span className="text-sm">Spotify</span>
                </div>
              </div>
              
              <div className="flex justify-center">
                <Button 
                  onClick={handleGeneratePlaylist} 
                  disabled={isLoading}
                  className="bg-gradient-to-r from-blue-500 to-purple-500 hover:from-blue-600 hover:to-purple-600 text-white font-semibold py-6 px-8 text-lg rounded-xl transition-all duration-200 transform hover:scale-[1.02] disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isLoading ? (
                    <div className="flex items-center justify-center space-x-2">
                      <Loader2 className="h-5 w-5 animate-spin" />
                      <span>Generating Playlist...</span>
                    </div>
                  ) : (
                    <div className="flex items-center justify-center space-x-2">
                      <Music className="h-5 w-5" />
                      <span>Generate My Playlist</span>
                    </div>
                  )}
                </Button>
              </div>
              
              <p className="text-sm text-gray-400">
                Your playlist will be created based on your Last.fm listening history from the past 30 days
              </p>
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
