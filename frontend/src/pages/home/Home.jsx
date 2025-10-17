import React, { useEffect, useState } from 'react';
import {API_URL} from "../../config";

function Home() {
  const [content, setContent] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${API_URL}/v0/content/home`)
      .then(response => response.json())
      .then(data => {
        setContent({
          title: data.title,
          html: data.html,
          version: data.version
        });
        setLoading(false);
      })
      .catch(error => {
        console.error('Failed to load homepage content:', error);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className='min-h-[calc(100vh-6rem)] bg-primary-background text-custom-gray-verylight flex flex-col justify-center py-8 px-4 sm:px-6 lg:px-8'>
        <div className='max-w-4xl mx-auto w-full text-center'>
          <p className='text-lg'>Loading...</p>
        </div>
      </div>
    );
  }

  if (!content) {
    return (
      <div className='min-h-[calc(100vh-6rem)] bg-primary-background text-custom-gray-verylight flex flex-col justify-center py-8 px-4 sm:px-6 lg:px-8'>
        <div className='max-w-4xl mx-auto w-full text-center'>
          <p className='text-lg text-red-400'>Failed to load homepage content.</p>
        </div>
      </div>
    );
  }

  return (
    <div className='min-h-[calc(100vh-6rem)] bg-primary-background text-custom-gray-verylight flex flex-col justify-center py-8 px-4 sm:px-6 lg:px-8'>
      <div className='max-w-4xl mx-auto w-full'>
        <div
          className='homepage-content'
          dangerouslySetInnerHTML={{ __html: content.html }}
        />
      </div>
    </div>
  );
}

export default Home;
