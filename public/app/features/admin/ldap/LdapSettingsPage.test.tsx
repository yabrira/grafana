import { screen, render, waitFor } from 'test/test-utils';

import LdapSettingsPage from './LdapSettingsPage';

const emptyLdapPayload = {
  id: 'ldap',
  provider: 'ldap',
  source: 'db',
  settings: {
    activeSyncEnabled: false,
    allowSignUp: false,
    config: {
      servers: [
        {
          attributes: {},
          bind_dn: '',
          bind_password: '',
          client_cert: '',
          client_cert_value: '',
          client_key: '',
          client_key_value: '',
          group_mappings: [],
          group_search_base_dns: [],
          group_search_filter: '',
          group_search_filter_user_attribute: '',
          host: 'ldap.example.com',
          min_tls_version: '',
          port: 389,
          root_ca_cert: '',
          root_ca_cert_value: [],
          search_base_dns: [],
          search_filter: '',
          skip_org_role_sync: false,
          ssl_skip_verify: false,
          start_tls: false,
          timeout: 10,
          tls_ciphers: [],
          tls_skip_verify: false,
          use_ssl: false,
        },
      ],
    },
    enabled: false,
    skipOrgRoleSync: false,
    syncCron: '',
  },
};

jest.mock('@grafana/runtime', () => ({
  ...jest.requireActual('@grafana/runtime'),
  getBackendSrv: () => ({
    get: jest.fn().mockResolvedValue(emptyLdapPayload),
    put: jest.fn(),
    delete: jest.fn(),
  }),
  reportInteraction: jest.fn(),
}));

jest.mock('app/core/components/FormPrompt/FormPrompt', () => ({
  FormPrompt: () => null,
}));

describe('LdapSettingsPage characterization', () => {
  it('renders via react-redux connect without throwing', async () => {
    render(<LdapSettingsPage />, {
      preloadedState: {
        navIndex: {
          LDAP: { id: 'LDAP', text: 'LDAP', url: '/admin/authentication/ldap' },
        },
      },
    });

    await waitFor(() => {
      expect(screen.queryByText(/loading/i)).not.toBeInTheDocument();
    });

    expect(await screen.findByDisplayValue('ldap.example.com')).toBeInTheDocument();
  });
});
