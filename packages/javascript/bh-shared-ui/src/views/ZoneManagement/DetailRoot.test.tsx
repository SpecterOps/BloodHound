import { render } from '../../test-utils';
import { apiClient } from '../../utils';
import DetailsRoot from './DetailsRoot';

const assetGroupSpy = vi.spyOn(apiClient, 'getAssetGroupTags');

describe('DetailsRoot', () => {
    it('calls getAssetGroupTags for topTagId', () => {
        render(<DetailsRoot />);
        expect(assetGroupSpy).toHaveBeenCalled();
    });
});
