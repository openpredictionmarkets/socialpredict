import React, { useEffect, useMemo, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import {
  listAdminUsers,
  promoteUserToModerator,
  updateModeratorSuspension,
} from '../../../api/adminUsersApi';

const roleBadgeClass = (user) => {
  if (user.usertype === 'ADMIN') return 'bg-sky-700 text-sky-50';
  if (user.usertype === 'MODERATOR' && user.moderatorStatus === 'suspended') return 'bg-rose-700 text-rose-50';
  if (user.usertype === 'MODERATOR') return 'bg-emerald-700 text-emerald-50';
  return 'bg-gray-700 text-gray-100';
};

const moderatorLabel = (user) => {
  if (user.usertype !== 'MODERATOR') return 'not approved';
  return user.moderatorStatus || 'active';
};

function UserQueue() {
  const { token } = useAuth();
  const [users, setUsers] = useState([]);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [pendingUsername, setPendingUsername] = useState('');
  const [reasonByUsername, setReasonByUsername] = useState({});

  const counts = useMemo(() => users.reduce((acc, user) => {
    if (user.usertype === 'MODERATOR') {
      acc.moderators += 1;
      if (user.moderatorStatus === 'suspended') {
        acc.suspended += 1;
      }
    } else if (user.usertype !== 'ADMIN') {
      acc.candidates += 1;
    }
    return acc;
  }, { candidates: 0, moderators: 0, suspended: 0 }), [users]);

  const loadUsers = async () => {
    setError('');
    setIsLoading(true);
    try {
      const result = await listAdminUsers({ token });
      setUsers(result.users || []);
    } catch (err) {
      setError(err.message || 'Failed to load user queue.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadUsers();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const reasonFor = (username) => reasonByUsername[username] || '';

  const setReasonFor = (username, reason) => {
    setReasonByUsername((current) => ({
      ...current,
      [username]: reason,
    }));
  };

  const updateUserInQueue = (updatedUser) => {
    setUsers((current) => current.map((user) => (
      user.username === updatedUser.username ? updatedUser : user
    )));
  };

  const runUserAction = async (username, action) => {
    setError('');
    setPendingUsername(username);
    try {
      const reason = reasonFor(username);
      let updatedUser;
      if (action === 'promote') {
        updatedUser = await promoteUserToModerator({ token, username, reason });
      } else if (action === 'suspend') {
        updatedUser = await updateModeratorSuspension({ token, username, suspended: true, reason });
      } else {
        updatedUser = await updateModeratorSuspension({ token, username, suspended: false, reason });
      }
      updateUserInQueue(updatedUser);
      setReasonFor(username, '');
    } catch (err) {
      setError(err.message || 'User action failed.');
    } finally {
      setPendingUsername('');
    }
  };

  return (
    <section className="p-6 bg-primary-background shadow-md rounded-lg text-white">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">
            Moderator governance
          </p>
          <h1 className="mt-2 text-2xl font-bold">User Queue</h1>
          <p className="mt-2 max-w-3xl text-sm text-gray-300">
            Review all users, see who is approved as a moderator, and perform
            baseline moderator promotion or suspension actions.
          </p>
        </div>
        <button
          type="button"
          onClick={loadUsers}
          disabled={isLoading}
          className="rounded-md bg-primary-pink px-4 py-2 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
        >
          {isLoading ? 'Refreshing...' : 'Refresh Queue'}
        </button>
      </div>

      <div className="mt-5 grid grid-cols-1 gap-3 sm:grid-cols-3">
        <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
          <p className="text-xs uppercase tracking-wide text-gray-400">Candidates</p>
          <p className="mt-1 text-2xl font-bold">{counts.candidates}</p>
        </div>
        <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
          <p className="text-xs uppercase tracking-wide text-gray-400">Moderators</p>
          <p className="mt-1 text-2xl font-bold">{counts.moderators}</p>
        </div>
        <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
          <p className="text-xs uppercase tracking-wide text-gray-400">Suspended</p>
          <p className="mt-1 text-2xl font-bold">{counts.suspended}</p>
        </div>
      </div>

      {error && (
        <div className="mt-5 rounded-md bg-red-700 p-3 text-sm text-white">
          {error}
        </div>
      )}

      <div className="mt-6 overflow-x-auto rounded-lg border border-gray-700">
        <table className="min-w-full divide-y divide-gray-700 text-left text-sm">
          <thead className="bg-gray-900 text-xs uppercase tracking-wide text-gray-300">
            <tr>
              <th className="px-4 py-3">User</th>
              <th className="px-4 py-3">Role</th>
              <th className="px-4 py-3">Moderator</th>
              <th className="px-4 py-3">Reason</th>
              <th className="px-4 py-3">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-800 bg-gray-950/30">
            {users.map((user) => {
              const isModerator = user.usertype === 'MODERATOR';
              const isSuspended = isModerator && user.moderatorStatus === 'suspended';
              const isAdmin = user.usertype === 'ADMIN';
              const pending = pendingUsername === user.username;

              return (
                <tr key={user.username}>
                  <td className="px-4 py-4 align-top">
                    <div className="font-semibold text-white">{user.displayName || user.username}</div>
                    <div className="font-mono text-xs text-gray-400">{user.username}</div>
                  </td>
                  <td className="px-4 py-4 align-top">
                    <span className={`rounded-full px-2 py-1 text-xs font-semibold ${roleBadgeClass(user)}`}>
                      {user.usertype}
                    </span>
                  </td>
                  <td className="px-4 py-4 align-top">
                    <div className="font-semibold">{moderatorLabel(user)}</div>
                    {user.moderatorSuspensionReason && (
                      <div className="mt-1 text-xs text-gray-400">
                        {user.moderatorSuspensionReason}
                      </div>
                    )}
                    {user.moderatorSuspendedAt && (
                      <div className="mt-1 text-xs text-gray-500">
                        Suspended {new Date(user.moderatorSuspendedAt).toLocaleString()}
                      </div>
                    )}
                  </td>
                  <td className="px-4 py-4 align-top">
                    <input
                      type="text"
                      value={reasonFor(user.username)}
                      onChange={(event) => setReasonFor(user.username, event.target.value)}
                      placeholder="Reason for audit log"
                      className="w-56 rounded-md border border-gray-700 bg-gray-900 px-3 py-2 text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
                    />
                  </td>
                  <td className="px-4 py-4 align-top">
                    <div className="flex flex-wrap gap-2">
                      {!isAdmin && !isModerator && (
                        <button
                          type="button"
                          disabled={pending}
                          onClick={() => runUserAction(user.username, 'promote')}
                          className="rounded-md bg-emerald-700 px-3 py-2 text-xs font-semibold text-white hover:bg-emerald-600 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          Approve Moderator
                        </button>
                      )}
                      {isModerator && !isSuspended && (
                        <button
                          type="button"
                          disabled={pending || !reasonFor(user.username).trim()}
                          onClick={() => runUserAction(user.username, 'suspend')}
                          className="rounded-md bg-rose-700 px-3 py-2 text-xs font-semibold text-white hover:bg-rose-600 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          Suspend
                        </button>
                      )}
                      {isModerator && isSuspended && (
                        <button
                          type="button"
                          disabled={pending}
                          onClick={() => runUserAction(user.username, 'unsuspend')}
                          className="rounded-md bg-sky-700 px-3 py-2 text-xs font-semibold text-white hover:bg-sky-600 disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          Unsuspend
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              );
            })}
            {!isLoading && users.length === 0 && (
              <tr>
                <td className="px-4 py-8 text-center text-gray-400" colSpan="5">
                  No users found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </section>
  );
}

export default UserQueue;
