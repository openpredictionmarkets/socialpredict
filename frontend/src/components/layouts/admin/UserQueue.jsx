import React, { useEffect, useMemo, useState } from 'react';
import { useAuth } from '../../../helpers/AuthContent';
import {
  listAdminUsers,
  promoteUserToModerator,
  updateModeratorSuspension,
} from '../../../api/adminUsersApi';
import SiteTabs from '../../tabs/SiteTabs';
import AdminAddUser from './AddUser';

const userQueueTabs = [
  {
    label: 'Non-Moderators',
    usertype: 'REGULAR',
    emptyMessage: 'No non-moderator users found.',
    searchPlaceholder: 'Search non-moderators by username, display name, or email',
  },
  {
    label: 'Moderators',
    usertype: 'MODERATOR',
    emptyMessage: 'No moderators found.',
    searchPlaceholder: 'Search moderators by username, display name, or email',
  },
];

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

const userMatchesTab = (user, usertype) => user?.usertype === usertype;

const UserQueueTab = ({ config }) => {
  const { token } = useAuth();
  const [users, setUsers] = useState([]);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [pendingUsername, setPendingUsername] = useState('');
  const [reasonByUsername, setReasonByUsername] = useState({});
  const [searchQuery, setSearchQuery] = useState('');

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

  const loadUsers = async (query = searchQuery) => {
    setError('');
    setIsLoading(true);
    try {
      const result = await listAdminUsers({
        token,
        usertype: config.usertype,
        query,
        limit: 250,
      });
      setUsers(result.users || []);
    } catch (err) {
      setError(err.message || 'Failed to load user queue.');
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (!token) {
      return undefined;
    }

    const timeoutId = window.setTimeout(() => {
      loadUsers(searchQuery);
    }, 300);

    return () => {
      window.clearTimeout(timeoutId);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token, config.usertype, searchQuery]);

  const reasonFor = (username) => reasonByUsername[username] || '';

  const setReasonFor = (username, reason) => {
    setReasonByUsername((current) => ({
      ...current,
      [username]: reason,
    }));
  };

  const updateUserInQueue = (updatedUser) => {
    setUsers((current) => {
      if (!userMatchesTab(updatedUser, config.usertype)) {
        return current.filter((user) => user.username !== updatedUser.username);
      }
      return current.map((user) => (
        user.username === updatedUser.username ? updatedUser : user
      ));
    });
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
    <div className="grid gap-5">
      <div className="grid gap-2 rounded-lg border border-gray-700 bg-gray-900/70 p-4">
        <label htmlFor={`user-queue-search-${config.usertype}`} className="text-xs font-mono uppercase tracking-[0.16em] text-gray-400">
          Search users
        </label>
        <div className="relative">
          <input
            id={`user-queue-search-${config.usertype}`}
            type="search"
            value={searchQuery}
            onChange={(event) => setSearchQuery(event.target.value)}
            placeholder={config.searchPlaceholder}
            className="w-full rounded-md border border-gray-600 bg-gray-800 px-3 py-2 pr-10 text-sm text-white focus:border-primary-pink focus:outline-none focus:ring-2 focus:ring-primary-pink/40"
          />
          {isLoading && (
            <div className="absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 animate-spin rounded-full border-b-2 border-primary-pink" />
          )}
        </div>
      </div>

      {config.usertype === 'REGULAR' && (
        <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
          <p className="text-xs uppercase tracking-wide text-gray-400">Non-moderator users</p>
          <p className="mt-1 text-2xl font-bold">{counts.candidates}</p>
        </div>
      )}
      {config.usertype === 'MODERATOR' && (
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
            <p className="text-xs uppercase tracking-wide text-gray-400">Moderators</p>
            <p className="mt-1 text-2xl font-bold">{counts.moderators}</p>
          </div>
          <div className="rounded-md border border-gray-700 bg-gray-900 p-3">
            <p className="text-xs uppercase tracking-wide text-gray-400">Suspended</p>
            <p className="mt-1 text-2xl font-bold">{counts.suspended}</p>
          </div>
        </div>
      )}

      {error && (
        <div className="rounded-md bg-red-700 p-3 text-sm text-white">
          {error}
        </div>
      )}

      <div className="overflow-x-auto rounded-lg border border-gray-700">
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
                  {config.emptyMessage}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

function UserQueue({ defaultTab = 'Non-Moderators' }) {
  const tabsData = [
    ...userQueueTabs.map((tab) => ({
      label: tab.label,
      content: <UserQueueTab config={tab} />,
    })),
    {
      label: 'Add User',
      content: <AdminAddUser />,
    },
  ];

  return (
    <section className="p-6 bg-primary-background shadow-md rounded-lg text-white">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.22em] text-primary-pink">
            Moderator governance
          </p>
          <h1 className="mt-2 text-2xl font-bold">User Governance</h1>
          <p className="mt-2 max-w-3xl text-sm text-gray-300">
            Review users by moderator status and perform baseline moderator promotion or suspension actions.
          </p>
        </div>
      </div>

      <div className="mt-6">
        <SiteTabs tabs={tabsData} defaultTab={defaultTab} />
      </div>
    </section>
  );
}

export default UserQueue;
