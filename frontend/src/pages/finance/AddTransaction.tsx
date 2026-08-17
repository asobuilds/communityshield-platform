import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function AddTransaction() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    amount: '',
    type: 'donation',
    description: '',
    transactionDate: '',
    paymentMethod: 'cash',
  });

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const userStr = localStorage.getItem('user');
      if (!userStr) {
        toast.error('Please login');
        navigate('/login');
        return;
      }
      const user = JSON.parse(userStr);
      const unitId = '00000000-0000-0000-0000-000000000000';
      const payload = {
        unitId,
        amount: parseFloat(formData.amount),
        type: formData.type,
        description: formData.description,
        transactionDate: formData.transactionDate || new Date().toISOString().split('T')[0],
        paymentMethod: formData.paymentMethod,
        initiatedBy: user.id,
      };
      await axios.post('/api/v1/transactions', payload);
      toast.success('Transaction added');
      navigate('/ledger');
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-2xl mx-auto bg-white rounded-lg shadow p-6">
        <h1 className="text-3xl font-bold text-blue-600 mb-6">💰 Add Transaction</h1>
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Type</label>
            <select name="type" value={formData.type} onChange={handleChange} className="w-full px-3 py-2 border rounded-lg">
              <option value="donation">Donation</option>
              <option value="bail">Bail</option>
              <option value="tax">Tax</option>
              <option value="gift">Gift</option>
              <option value="expense">Expense</option>
            </select>
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Amount (₦)</label>
            <input type="number" name="amount" value={formData.amount} onChange={handleChange} className="w-full px-3 py-2 border rounded-lg" required step="0.01" />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Description</label>
            <textarea name="description" value={formData.description} onChange={handleChange} rows={3} className="w-full px-3 py-2 border rounded-lg" required />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Date</label>
            <input type="date" name="transactionDate" value={formData.transactionDate} onChange={handleChange} className="w-full px-3 py-2 border rounded-lg" />
          </div>
          <div className="mb-4">
            <label className="block text-sm font-medium mb-1">Payment Method</label>
            <select name="paymentMethod" value={formData.paymentMethod} onChange={handleChange} className="w-full px-3 py-2 border rounded-lg">
              <option value="cash">Cash</option>
              <option value="bank_transfer">Bank Transfer</option>
              <option value="mobile_money">Mobile Money</option>
            </select>
          </div>
          <button type="submit" disabled={loading} className="w-full bg-green-600 text-white py-2 rounded hover:bg-green-700 disabled:opacity-50">
            {loading ? 'Adding...' : 'Add Transaction'}
          </button>
        </form>
        <div className="mt-4 text-center">
          <a href="/finance" className="text-blue-600 hover:underline">← Back</a>
        </div>
      </div>
    </div>
  );
}

export default AddTransaction;