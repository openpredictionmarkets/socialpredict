import React, { useState, useEffect } from 'react';
import { API_URL } from '../../config';
import { useAuth } from '../../helpers/AuthContent';

const PollCard = ({ poll, token, onUpdate }) => {
  const { username } = useAuth();
  const [voting, setVoting] = useState(false);
  const [closing, setClosing] = useState(false);
  const [error, setError] = useState('');

  const total = poll.yesCount + poll.noCount;
  const yesPct = total > 0 ? Math.round((poll.yesCount / total) * 100) : 50;
  const noPct = total > 0 ? 100 - yesPct : 50;

  const vote = async (choice) => {
    setVoting(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/v0/polls/${poll.id}/vote`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ vote: choice }),
      });
      if (!res.ok) {
        const msg = await res.text();
        throw new Error(msg || `HTTP ${res.status}`);
      }
      const updated = await res.json();
      onUpdate(updated);
    } catch (err) {
      setError(err.message);
    } finally {
      setVoting(false);
    }
  };

  const closePoll = async () => {
    setClosing(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/v0/polls/${poll.id}/close`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const msg = await res.text();
        throw new Error(msg || `HTTP ${res.status}`);
      }
      const updated = await res.json();
      onUpdate(updated);
    } catch (err) {
      setError(err.message);
    } finally {
      setClosing(false);
    }
  };

  const canVote = token && !poll.isClosed && !poll.userVote;
  const isCreator = username === poll.creatorUsername;

  return (
    <div className='bg-gray-700 rounded-lg p-4 space-y-3'>
      <div className='flex items-start justify-between gap-2'>
        <div className='flex-1'>
          <p className='text-gray-200 font-medium'>{poll.question}</p>
          {poll.description && (
            <p className='text-gray-400 text-sm mt-0.5'>{poll.description}</p>
          )}
          <p className='text-gray-500 text-xs mt-1'>
            by @{poll.creatorUsername} · {total} vote{total !== 1 ? 's' : ''}
          </p>
        </div>
        {poll.isClosed && (
          <span className='text-xs bg-gray-600 text-gray-300 px-2 py-0.5 rounded shrink-0'>
            Closed
          </span>
        )}
      </div>

      {/* Results bar */}
      <div className='space-y-1'>
        <div className='flex justify-between text-xs text-gray-400'>
          <span>YES — {poll.yesCount} ({yesPct}%)</span>
          <span>NO — {poll.noCount} ({noPct}%)</span>
        </div>
        <div className='h-2 bg-gray-600 rounded-full overflow-hidden'>
          <div
            className='h-full bg-green-500 rounded-full transition-all duration-300'
            style={{ width: `${yesPct}%` }}
          />
        </div>
      </div>

      {/* Vote buttons */}
      {canVote && (
        <div className='flex gap-2'>
          <button
            onClick={() => vote('YES')}
            disabled={voting}
            className='flex-1 py-1.5 bg-green-600 hover:bg-green-500 text-white text-sm rounded-lg
                       disabled:opacity-50 disabled:cursor-not-allowed transition-colors'
          >
            Vote YES
          </button>
          <button
            onClick={() => vote('NO')}
            disabled={voting}
            className='flex-1 py-1.5 bg-red-600 hover:bg-red-500 text-white text-sm rounded-lg
                       disabled:opacity-50 disabled:cursor-not-allowed transition-colors'
          >
            Vote NO
          </button>
        </div>
      )}

      {poll.userVote && (
        <p className='text-sm text-gray-400'>
          You voted <span className='font-semibold text-gray-200'>{poll.userVote}</span>
        </p>
      )}

      {!token && !poll.isClosed && (
        <p className='text-xs text-gray-500'>Log in to vote.</p>
      )}

      {error && <p className='text-red-400 text-xs'>{error}</p>}

      {isCreator && !poll.isClosed && (
        <button
          onClick={closePoll}
          disabled={closing}
          className='text-xs text-gray-400 hover:text-gray-200 underline disabled:opacity-50'
        >
          {closing ? 'Closing…' : 'Close poll'}
        </button>
      )}
    </div>
  );
};

const CreatePollForm = ({ token, onCreate }) => {
  const [question, setQuestion] = useState('');
  const [description, setDescription] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const submit = async (e) => {
    e.preventDefault();
    if (!question.trim()) return;
    setSubmitting(true);
    setError('');
    try {
      const res = await fetch(`${API_URL}/v0/polls`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${token}` },
        body: JSON.stringify({ question: question.trim(), description: description.trim() }),
      });
      if (!res.ok) {
        const msg = await res.text();
        throw new Error(msg || `HTTP ${res.status}`);
      }
      const poll = await res.json();
      onCreate(poll);
      setQuestion('');
      setDescription('');
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={submit} className='bg-gray-700 rounded-lg p-4 space-y-3'>
      <p className='text-gray-200 font-medium'>Create a Poll</p>
      <input
        type='text'
        placeholder='Ask a yes/no question…'
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
        maxLength={200}
        className='w-full bg-gray-800 text-gray-200 rounded-lg px-3 py-2 text-sm
                   border border-gray-600 focus:outline-none focus:border-blue-500'
      />
      <textarea
        placeholder='Description (optional)'
        value={description}
        onChange={(e) => setDescription(e.target.value)}
        maxLength={2000}
        rows={2}
        className='w-full bg-gray-800 text-gray-200 rounded-lg px-3 py-2 text-sm
                   border border-gray-600 focus:outline-none focus:border-blue-500 resize-none'
      />
      {error && <p className='text-red-400 text-xs'>{error}</p>}
      <button
        type='submit'
        disabled={submitting || !question.trim()}
        className='px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm rounded-lg
                   disabled:opacity-50 disabled:cursor-not-allowed transition-colors'
      >
        {submitting ? 'Creating…' : 'Create Poll'}
      </button>
    </form>
  );
};

function Polls() {
  const [polls, setPolls] = useState([]);
  const [loading, setLoading] = useState(true);
  const { username } = useAuth();
  const token = localStorage.getItem('token');

  useEffect(() => {
    fetch(`${API_URL}/v0/polls`, {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then((r) => r.json())
      .then((data) => setPolls(Array.isArray(data) ? data : []))
      .catch(() => setPolls([]))
      .finally(() => setLoading(false));
  }, []);

  const handleUpdate = (updated) => {
    setPolls((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
  };

  const handleCreate = (newPoll) => {
    setPolls((prev) => [newPoll, ...prev]);
  };

  return (
    <div className='p-6 space-y-4 max-w-2xl mx-auto'>
      <h1 className='text-xl font-semibold text-gray-200'>Polls</h1>
      <p className='text-gray-400 text-sm'>
        Quick yes/no questions from the community.
      </p>

      {username && token && (
        <CreatePollForm token={token} onCreate={handleCreate} />
      )}

      {loading ? (
        <p className='text-gray-400 text-sm'>Loading polls…</p>
      ) : polls.length === 0 ? (
        <p className='text-gray-400 text-sm'>No open polls yet.</p>
      ) : (
        polls.map((poll) => (
          <PollCard
            key={poll.id}
            poll={poll}
            token={token}
            onUpdate={handleUpdate}
          />
        ))
      )}
    </div>
  );
}

export default Polls;
