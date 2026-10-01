import { screen, render } from 'test/test-utils';

import UpgradePage from './UpgradePage';

jest.mock('./ServerStats', () => ({
  ServerStats: () => <div data-testid="server-stats-mock" />,
}));

describe('UpgradePage characterization', () => {
  it('renders via react-redux connect without throwing', async () => {
    render(<UpgradePage />, {
      preloadedState: {
        navIndex: {
          upgrading: { id: 'upgrading', text: 'Upgrading', url: '/admin/upgrading' },
        },
      },
    });

    expect(await screen.findByRole('heading', { name: /enterprise license/i })).toBeInTheDocument();
    expect(screen.getByTestId('server-stats-mock')).toBeInTheDocument();
  });
});
