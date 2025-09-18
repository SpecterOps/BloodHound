import { Card, CardContent, CardFooter, CardTitle } from '@bloodhoundenterprise/doodleui';
import { AppIcon } from '../../../components';
import { useHistoryTableContext } from './HistoryTableContext';

const HistoryNotes = () => {
    const { currentNote, showNoteDetails } = useHistoryTableContext();

    return (
        <div>
            <Card className='flex justify-center mb-4 p-4 h-[56px] w-[400px]  min-w-[300px] max-w-[32rem] '>
                <CardTitle className='flex  items-center gap-1'>
                    <AppIcon.LinedPaper size={18} />
                    Note
                </CardTitle>
            </Card>

            {showNoteDetails && (
                <Card className='p-4 '>
                    <CardContent>
                        <p>{currentNote?.note}</p>
                    </CardContent>
                    <CardFooter className='text-xs'>
                        <p>
                            By {currentNote?.createdBy} on {currentNote?.timestamp}
                        </p>
                    </CardFooter>
                </Card>
            )}
        </div>
    );
};

export default HistoryNotes;
