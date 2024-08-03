import React from 'react';
import { Link } from 'react-router-dom';

const NotFound = () => {
  return (
    <div className='flex items-center justify-center min-h-[calc(100vh-6rem)] bg-primary-background -mb-24'>
      <div className='text-center'>
        <h1 className='text-4xl font-bold mb-4'>404 - Page Not Found</h1>
        <p className='text-xl mb-8'>
          Oops! The page you're looking for doesn't exist.
        </p>
        <Link
          to='/'
          className='bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded transition duration-300'
        >
          Go Home
        </Link>
      </div>
    </div>
  );
};

export default NotFound;
