import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { render } from 'test/test-utils';

import { AppleSignInButton } from './AppleSignInButton';

describe('AppleSignInButton', () => {
  it('renders a button and not a link, so the stub has no route to navigate to', () => {
    render(<AppleSignInButton />);

    expect(screen.getByRole('button', { name: 'Sign in with Apple' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Sign in with Apple' })).not.toBeInTheDocument();
  });

  it('shows the not-configured notice only once the button is clicked', async () => {
    render(<AppleSignInButton />);

    expect(screen.queryByRole('status')).not.toBeInTheDocument();

    await userEvent.click(screen.getByRole('button', { name: 'Sign in with Apple' }));

    const notice = screen.getByRole('status', { name: 'Sign in with Apple is not available' });
    expect(notice).toHaveTextContent(
      'This sign-in method is not configured on this Grafana instance. Use one of the other sign-in options.'
    );
  });
});
