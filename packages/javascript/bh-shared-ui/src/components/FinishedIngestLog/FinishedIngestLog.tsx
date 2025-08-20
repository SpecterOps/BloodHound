import { Paper, Typography } from '@mui/material';
import { FC } from 'react';
import { FileIngestTable } from '../FileIngestTable';
import PageWithTitle from '../PageWithTitle';

const FinishedJobsLog: FC = () => {
    return (
        <PageWithTitle
            title='Finished Jobs Log'
            data-testid='finished-tasks-log'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Review completed collection jobs from Enterprise collectors here.
                </Typography>
            }>
            <Paper>
                <FileIngestTable />
            </Paper>
        </PageWithTitle>
    );
};

export default FinishedJobsLog;
