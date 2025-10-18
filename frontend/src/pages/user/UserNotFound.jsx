import React from 'react';
import { Link } from 'react-router-dom';

const UserNotFound = ({ username }) => {
  return (
    <div className="flex items-center justify-center min-h-screen bg-primary-background">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4 text-white">No User Found</h1>
        <p className="text-xl mb-6 text-gray-400">
          A user with the username "<span className="font-mono text-white">@{username}</span>" could not be found.
        </p>
        <p className="text-lg mb-8 text-gray-500">
          The user may have been deleted, or the username might be incorrect.
        </p>
        <div className="flex gap-4 justify-center">
          <Link
            to="/"
            className="bg-primary-pink hover:bg-primary-pink/80 text-white font-semibold py-2 px-6 rounded-lg transition-colors duration-200"
          >
            Go Home
          </Link>
          <Link
            to="/markets"
            className="bg-gray-700 hover:bg-gray-600 text-white font-semibold py-2 px-6 rounded-lg transition-colors duration-200"
          >
            Browse Markets
          </Link>
        </div>
      </div>
    </div>
  );
};

export default UserNotFound;
