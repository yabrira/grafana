import { Route, Routes } from 'react-router-dom-v5-compat';
import { screen, render } from 'test/test-utils';

import UserAdminPage from './UserAdminPage';

const mockUser = {
  id: 1,
  uid: 'user-1',
  login: 'admin',
  email: 'admin@localhost',
  name: 'Admin',
  isGrafanaAdmin: true,
  isDisabled: false,
  isExternal: false,
  isProvisioned: false,
  authLabels: [],
};

jest.mock('./state/actions', () => {
  const actual = jest.requireActual('./state/actions');
  return {
    ...actual,
    loadAdminUserPage: () => () => Promise.resolve(),
  };
});

jest.mock('app/core/services/context_srv', () => ({
  contextSrv: {
    user: { orgId: 1, timezone: 'browser', weekStart: 'monday' },
    hasPermission: () => true,
    hasPermissionInMetadata: () => true,
    licensedAccessControlEnabled: () => false,
  },
}));

describe('UserAdminPage characterization', () => {
  it('renders via react-redux connect without throwing', async () => {
    render(
      <Routes>
        <Route path="/admin/users/edit/:id" element={<UserAdminPage />} />
      </Routes>,
      {
        historyOptions: { initialEntries: ['/admin/users/edit/user-1'] },
        preloadedState: {
          navIndex: {
            'global-users': { id: 'global-users', text: 'Users', url: '/admin/users' },
          },
          userAdmin: {
            user: mockUser,
            sessions: [],
            orgs: [],
            isLoading: false,
            error: undefined,
          },
        },
      }
    );

    // Pin current connected render path: profile section for the preloaded user.
    expect(await screen.findByText(/manage settings for an individual user/i)).toBeInTheDocument();
    expect(screen.getByText('Admin')).toBeInTheDocument();
  });
});
