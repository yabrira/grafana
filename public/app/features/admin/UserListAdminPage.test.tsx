import { screen, render } from 'test/test-utils';

import { selectors as e2eSelectors } from '@grafana/e2e-selectors';

import UserListAdminPage from './UserListAdminPage';

const selectors = e2eSelectors.pages.UserListPage.UserListAdminPage;

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getBackendSrv: () => ({
    get: jest.fn().mockResolvedValue({ users: [], totalCount: 0, page: 1, perPage: 50 }),
  }),
}));

jest.mock('app/core/services/context_srv', () => ({
  contextSrv: {
    user: { orgId: 1, timezone: 'browser', weekStart: 'monday' },
    hasPermission: () => true,
    licensedAccessControlEnabled: () => false,
  },
}));

describe('UserListAdminPage characterization', () => {
  it('renders via react-redux connect without throwing', async () => {
    render(<UserListAdminPage />, {
      preloadedState: {
        navIndex: {
          'global-users': { id: 'global-users', text: 'Users', url: '/admin/users' },
        },
      },
    });

    expect(await screen.findByTestId(selectors.container)).toBeInTheDocument();
  });
});
