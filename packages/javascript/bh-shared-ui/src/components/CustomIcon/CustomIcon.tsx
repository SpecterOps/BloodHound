import { IconProp, Transform, library } from '@fortawesome/fontawesome-svg-core';
import { fas } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

library.add(fas);

interface CustomNodeIconProps {
    icon: IconProp;
    transform: string | Transform;
}

const CustomNodeIcon: React.FC<CustomNodeIconProps> = ({ icon, transform }) => {
    return <FontAwesomeIcon icon={icon} transform={transform} />;
};

export default CustomNodeIcon;
