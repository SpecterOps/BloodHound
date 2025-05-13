import { findIconDefinition } from '@fortawesome/fontawesome-svg-core';
import { IconName } from '@fortawesome/free-solid-svg-icons';
import { apiClient, IconDictionary } from 'bh-shared-ui';
import { SagaIterator } from 'redux-saga';
import { all, call, fork, put, takeEvery } from 'redux-saga/effects';
import { appendSvgUrls, DEFAULT_ICON_COLOR, NODE_SCALE } from 'src/views/Explore/svgIcons';
import { setCustomNodeInformation } from './actions';
import { GLOBAL_FETCH_CUSTOM_NODE_INFORMATION } from './types';

function* customNodeInfoWatcher(): SagaIterator {
    yield takeEvery([GLOBAL_FETCH_CUSTOM_NODE_INFORMATION], customNodeInfoWorker);
}

function* customNodeInfoWorker() {
    try {
        const response: unknown = yield call(apiClient.getCustomNodeKinds);
        const nodes: any[] = (response as any).data.data;

        // transform here, dispatch finished objects to store
        const customNodes = transformNodes(nodes);
        yield put(setCustomNodeInformation(customNodes));
    } catch (e) {
        yield call(console.error, 'Failure when attempting to fetch available asset groups');
        yield call(console.error, e);
    }
}

export default function* StartGlobalSagas() {
    yield all([fork(customNodeInfoWatcher)]);
}

function transformNodes(nodes: any[]): IconDictionary {
    const customIcons: IconDictionary = {};

    nodes.forEach((node) => {
        const iconName = node.config.icon.name as IconName;

        const iconDefinition = findIconDefinition({ prefix: 'fas', iconName: iconName });
        if (iconDefinition == undefined) {
            return;
        }

        customIcons[node.kindName] = {
            icon: iconDefinition,
            color: DEFAULT_ICON_COLOR,
        };
    });

    appendSvgUrls(customIcons, NODE_SCALE);

    return customIcons;
}
