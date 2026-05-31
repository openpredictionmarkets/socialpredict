const AUTH_STORAGE_KEYS = {
  token: 'token',
  username: 'username',
  usertype: 'usertype',
  changePasswordNeeded: 'changePasswordNeeded',
};

const read = (key) => localStorage.getItem(key);
const write = (key, value) => {
  if (value === undefined || value === null) {
    localStorage.removeItem(key);
    return;
  }

  localStorage.setItem(key, value);
};

export const authStorage = {
  getToken: () => read(AUTH_STORAGE_KEYS.token),
  getUsername: () => read(AUTH_STORAGE_KEYS.username),
  getUsertype: () => read(AUTH_STORAGE_KEYS.usertype),
  setUsername: (username) => write(AUTH_STORAGE_KEYS.username, username),
  saveLogin: ({ token, username, usertype }) => {
    write(AUTH_STORAGE_KEYS.token, token);
    write(AUTH_STORAGE_KEYS.username, username);
    write(AUTH_STORAGE_KEYS.usertype, usertype);
  },
  clearLegacyPasswordChangeFlag: () => {
    localStorage.removeItem(AUTH_STORAGE_KEYS.changePasswordNeeded);
  },
  clear: () => {
    Object.values(AUTH_STORAGE_KEYS).forEach((key) => localStorage.removeItem(key));
  },
};
