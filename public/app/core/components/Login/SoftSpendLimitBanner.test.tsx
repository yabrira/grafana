import { render, screen } from 'test/test-utils';

import { SoftSpendLimitBanner, SOFT_SPEND_LIMIT_FIXTURE } from './SoftSpendLimitBanner';

describe('SoftSpendLimitBanner', () => {
  it('shows the fixture usage percent and SAR amounts as an awareness warning', () => {
    render(<SoftSpendLimitBanner />);

    const banner = screen.getByRole('alert', { name: 'Soft spend limit' });
    expect(banner).toBeInTheDocument();
    expect(banner).toHaveTextContent(`${SOFT_SPEND_LIMIT_FIXTURE.usedPercent}%`);
    expect(banner).toHaveTextContent('SAR 41,000 / SAR 50,000');
    expect(banner).toHaveTextContent('Payments still go through');
    expect(banner).toHaveTextContent('not a hard block');
  });
});
