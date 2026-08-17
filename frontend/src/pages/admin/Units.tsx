import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import toast from 'react-hot-toast';

function Units() {
  const navigate = useNavigate();
  const [units, setUnits] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [editingUnit, setEditingUnit] = useState<any>(null);
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    type: 'neighborhood_watch',
    latitude: '',
    longitude: '',
    coverageArea: '',
    contactPerson: '',
    contactPhone: '',
    contactEmail: '',
    registrationNumber: '',
  });

  useEffect(() => {
    fetchUnits();
  }, []);

  const fetchUnits = async () => {
    try {
      const response = await axios.get('/api/v1/admin/units');
      setUnits(response.data.units || []);
    } catch (error: any) {
      toast.error('Failed to load units');
    } finally {
      setLoading(false);
    }
  };

  const handleFormChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleCreateUnit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await axios.post('/api/v1/admin/units', {
        ...formData,
        latitude: parseFloat(formData.latitude) || 0,
        longitude: parseFloat(formData.longitude) || 0,
      });
      toast.success('Security unit created');
      setShowCreateForm(false);
      resetForm();
      fetchUnits();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to create');
    }
  };

  const handleEditClick = (unit: any) => {
    setEditingUnit(unit);
    setFormData({
      name: unit.name,
      type: unit.type,
      latitude: unit.latitude?.toString() || '',
      longitude: unit.longitude?.toString() || '',
      coverageArea: unit.coverageArea || '',
      contactPerson: unit.contactPerson || '',
      contactPhone: unit.contactPhone || '',
      contactEmail: unit.contactEmail || '',
      registrationNumber: unit.registrationNumber || '',
    });
  };

  const handleUpdateUnit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingUnit) return;
    try {
      await axios.put(`/api/v1/admin/units/${editingUnit.id}`, {
        ...formData,
        latitude: parseFloat(formData.latitude) || 0,
        longitude: parseFloat(formData.longitude) || 0,
      });
      toast.success('Security unit updated');
      setEditingUnit(null);
      resetForm();
      fetchUnits();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to update');
    }
  };

  const handleDeleteUnit = async (id: string) => {
    if (!confirm('Delete this security unit?')) return;
    try {
      await axios.delete(`/api/v1/admin/units/${id}`);
      toast.success('Security unit deleted');
      fetchUnits();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to delete');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'neighborhood_watch',
      latitude: '',
      longitude: '',
      coverageArea: '',
      contactPerson: '',
      contactPhone: '',
      contactEmail: '',
      registrationNumber: '',
    });
  };

  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'vigilante': return 'Vigilante';
      case 'neighborhood_watch': return 'Neighborhood Watch';
      case 'community_police': return 'Community Police';
      default: return type;
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-between items-center mb-6 flex-wrap gap-2">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate(-1)}
              className="text-blue-600 hover:text-blue-800 font-medium"
            >
              ← Back
            </button>
            <h1 className="text-3xl font-bold text-blue-600">🏢 Security Units</h1>
          </div>
          <button
            onClick={() => setShowCreateForm(!showCreateForm)}
            className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
          >
            {showCreateForm ? 'Cancel' : '+ Create Security Unit'}
          </button>
        </div>

        {/* Create Form */}
        {showCreateForm && (
          <div className="bg-white rounded-lg shadow p-6 mb-6">
            <h2 className="text-xl font-semibold mb-4">Create Security Unit</h2>
            <form onSubmit={handleCreateUnit} className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name *</label>
                <input type="text" name="name" value={formData.name} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" required />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Type *</label>
                <select name="type" value={formData.type} onChange={handleFormChange} className="w-full px-3 py-2 border rounded">
                  <option value="vigilante">Vigilante</option>
                  <option value="neighborhood_watch">Neighborhood Watch</option>
                  <option value="community_police">Community Police</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Coverage Area</label>
                <input type="text" name="coverageArea" value={formData.coverageArea} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Person</label>
                <input type="text" name="contactPerson" value={formData.contactPerson} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Phone</label>
                <input type="text" name="contactPhone" value={formData.contactPhone} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Email</label>
                <input type="email" name="contactEmail" value={formData.contactEmail} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div className="md:col-span-2">
                <button type="submit" className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700">Create</button>
              </div>
            </form>
          </div>
        )}

        {/* Edit Form */}
        {editingUnit && (
          <div className="bg-white rounded-lg shadow p-6 mb-6 border-l-4 border-blue-600">
            <h2 className="text-xl font-semibold mb-4">Edit Security Unit</h2>
            <form onSubmit={handleUpdateUnit} className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">Name *</label>
                <input type="text" name="name" value={formData.name} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" required />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Type *</label>
                <select name="type" value={formData.type} onChange={handleFormChange} className="w-full px-3 py-2 border rounded">
                  <option value="vigilante">Vigilante</option>
                  <option value="neighborhood_watch">Neighborhood Watch</option>
                  <option value="community_police">Community Police</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Coverage Area</label>
                <input type="text" name="coverageArea" value={formData.coverageArea} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Person</label>
                <input type="text" name="contactPerson" value={formData.contactPerson} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Phone</label>
                <input type="text" name="contactPhone" value={formData.contactPhone} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Contact Email</label>
                <input type="email" name="contactEmail" value={formData.contactEmail} onChange={handleFormChange} className="w-full px-3 py-2 border rounded" />
              </div>
              <div className="md:col-span-2 flex gap-3">
                <button type="submit" className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700">Update</button>
                <button type="button" onClick={() => { setEditingUnit(null); resetForm(); }} className="bg-gray-600 text-white px-6 py-2 rounded hover:bg-gray-700">Cancel</button>
              </div>
            </form>
          </div>
        )}

        {/* Units List */}
        {loading ? (
          <div className="bg-white p-6 rounded shadow text-center">Loading...</div>
        ) : units.length === 0 ? (
          <div className="bg-white p-12 rounded shadow text-center">
            <p className="text-gray-500 text-lg">No security units found.</p>
            <button onClick={() => setShowCreateForm(true)} className="mt-4 bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">Create First Security Unit</button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {units.map((unit) => (
              <div key={unit.id} className="bg-white rounded-lg shadow p-6 border border-gray-100">
                <h3 className="text-xl font-semibold text-blue-600">{unit.name}</h3>
                <p className="text-sm text-gray-500">Type: {getTypeLabel(unit.type)}</p>
                {unit.coverageArea && <p className="text-sm text-gray-600">Area: {unit.coverageArea}</p>}
                {unit.contactPerson && <p className="text-sm text-gray-600">Contact: {unit.contactPerson}</p>}
                {unit.contactPhone && <p className="text-sm text-gray-600">Phone: {unit.contactPhone}</p>}
                <div className="mt-3 flex gap-2 flex-wrap">
                  <button onClick={() => handleEditClick(unit)} className="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700">Edit</button>
                  <button onClick={() => handleDeleteUnit(unit.id)} className="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">Delete</button>
                </div>
              </div>
            ))}
          </div>
        )}
        <div className="mt-6">
          <a href="/admin" className="text-blue-600 hover:underline">← Back to Admin Dashboard</a>
        </div>
      </div>
    </div>
  );
}

export default Units;