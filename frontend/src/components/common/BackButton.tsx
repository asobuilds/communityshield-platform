import { useNavigate } from 'react-router-dom';

interface BackButtonProps {
  to?: string;  // optional custom path
  label?: string;
}

function BackButton({ to, label = '← Back' }: BackButtonProps) {
  const navigate = useNavigate();
  const handleClick = () => {
    if (to) {
      navigate(to);
    } else {
      navigate(-1); // go back one page in history
    }
  };
  return (
    <button
      onClick={handleClick}
      className="text-blue-600 hover:underline text-sm mb-4 inline-block"
    >
      {label}
    </button>
  );
}

export default BackButton;