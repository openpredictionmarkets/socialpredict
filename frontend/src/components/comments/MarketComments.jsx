import React, { useState, useEffect, useCallback } from 'react';
import { API_URL } from '../../config';

const MarketComments = ({ marketId, isLoggedIn, token }) => {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState('');
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const username = localStorage.getItem('username');

  const fetchComments = useCallback(async () => {
    try {
      const response = await fetch(`${API_URL}/v0/markets/${marketId}/comments`);
      if (!response.ok) throw new Error('Failed to fetch comments');
      const data = await response.json();
      setComments(data || []);
    } catch (err) {
      setError('Could not load comments.');
    } finally {
      setLoading(false);
    }
  }, [marketId]);

  useEffect(() => {
    fetchComments();
  }, [fetchComments]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!newComment.trim()) return;

    setSubmitting(true);
    setError(null);
    try {
      const response = await fetch(`${API_URL}/v0/markets/${marketId}/comments`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ content: newComment.trim() }),
      });
      if (!response.ok) {
        const msg = await response.text();
        throw new Error(msg || 'Failed to post comment');
      }
      setNewComment('');
      await fetchComments();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (commentId) => {
    try {
      const response = await fetch(
        `${API_URL}/v0/markets/${marketId}/comments/${commentId}`,
        {
          method: 'DELETE',
          headers: { Authorization: `Bearer ${token}` },
        }
      );
      if (!response.ok) throw new Error('Failed to delete comment');
      setComments((prev) => prev.filter((c) => c.id !== commentId));
    } catch (err) {
      setError(err.message);
    }
  };

  return (
    <div className='mt-8'>
      <h3 className='text-lg font-semibold text-gray-200 mb-4'>
        Comments ({comments.length})
      </h3>

      {error && (
        <p className='text-red-400 text-sm mb-3'>{error}</p>
      )}

      {loading ? (
        <p className='text-gray-400 text-sm'>Loading comments…</p>
      ) : comments.length === 0 ? (
        <p className='text-gray-400 text-sm'>
          No comments yet. {isLoggedIn ? 'Be the first!' : 'Log in to comment.'}
        </p>
      ) : (
        <ul className='space-y-3 mb-4'>
          {comments.map((c) => (
            <li
              key={c.id}
              className='bg-gray-700 rounded-lg p-3 flex justify-between items-start'
            >
              <div>
                <span className='text-blue-400 text-sm font-medium'>
                  {c.username}
                </span>
                <p className='text-gray-200 text-sm mt-1 whitespace-pre-wrap'>
                  {c.content}
                </p>
                <p className='text-gray-500 text-xs mt-1'>
                  {new Date(c.createdAt).toLocaleString()}
                </p>
              </div>
              {isLoggedIn && c.username === username && (
                <button
                  onClick={() => handleDelete(c.id)}
                  className='text-gray-500 hover:text-red-400 text-xs ml-3 shrink-0 transition-colors'
                  aria-label='Delete comment'
                >
                  ✕
                </button>
              )}
            </li>
          ))}
        </ul>
      )}

      {isLoggedIn && (
        <form onSubmit={handleSubmit} className='flex flex-col gap-2'>
          <textarea
            value={newComment}
            onChange={(e) => setNewComment(e.target.value)}
            placeholder='Add a comment…'
            maxLength={2000}
            rows={3}
            className='w-full bg-gray-700 text-gray-200 rounded-lg p-3 text-sm
                       border border-gray-600 focus:outline-none focus:border-blue-400
                       resize-none placeholder-gray-500'
          />
          <button
            type='submit'
            disabled={submitting || !newComment.trim()}
            className='self-end px-4 py-2 bg-blue-600 text-white text-sm rounded-lg
                       hover:bg-blue-500 disabled:opacity-50 disabled:cursor-not-allowed
                       transition-colors'
          >
            {submitting ? 'Posting…' : 'Post Comment'}
          </button>
        </form>
      )}
    </div>
  );
};

export default MarketComments;
