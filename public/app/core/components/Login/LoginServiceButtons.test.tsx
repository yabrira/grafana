import { screen } from '@testing-library/react';
import { render } from 'test/test-utils';

import { config } from '@grafana/runtime';

import { LoginServiceButtons } from './LoginServiceButtons';

const originalOauth = config.oauth;
const originalSamlEnabled = config.samlEnabled;

afterEach(() => {
  config.oauth = originalOauth;
  config.samlEnabled = originalSamlEnabled;
});

describe('LoginServiceButtons', () => {
  it('renders Sign in with Apple with href login/apple when no OAuth providers are configured', () => {
    config.oauth = {};
    config.samlEnabled = false;

    render(<LoginServiceButtons />);

    const appleLink = screen.getByRole('link', { name: 'Sign in with Apple' });
    expect(appleLink).toHaveAttribute('href', 'login/apple');
    expect(screen.queryByRole('link', { name: 'Sign in with Google' })).not.toBeInTheDocument();
  });
});
