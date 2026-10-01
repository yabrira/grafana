import { screen, render } from 'test/test-utils';

import { selectors as e2eSelectors } from '@grafana/e2e-selectors';

import UserListAnonymousDevicesPage from './UserListAnonymousPage';

const selectors = e2eSelectors.pages.UserListPage.UserListAdminPage;

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getBackendSrv: () => ({
    get: jest.fn().mockResolvedValue({ devices: [], totalCount: 0, page: 1, perPage: 50 }),
  }),
}));

describe('UserListAnonymousDevicesPage characterization', () => {
  it('renders via react-redux connect without throwing', async () => {
    render(<UserListAnonymousDevicesPage />, {
      preloadedState: {
        navIndex: {
          'anonymous-users': { id: 'anonymous-users', text: 'Anonymous users', url: '/admin/users/anonymous' },
        },
      },
    });

    expect(await screen.findByTestId(selectors.container)).toBeInTheDocument();
  });
});
