import { PersistedState } from 'runed';

export const NS = 'tether';

export const loggedIn = new PersistedState<boolean>(`${NS}:logged-in`, false);
export const lang = new PersistedState<string>(`${NS}:lang`, 'yaml');
export const env = new PersistedState<string>(`${NS}:env`, 'default');
