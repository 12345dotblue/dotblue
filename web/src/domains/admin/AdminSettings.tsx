import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Empty,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Statistic,
  Table,
  Tabs,
  Tag,
  Tree,
  Typography,
  message,
} from 'antd';
import {
  CopyOutlined,
  PlusOutlined,
  SettingOutlined,
  TeamOutlined,
  ApartmentOutlined,
  LinkOutlined,
} from '@ant-design/icons';
import axios from 'axios';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router-dom';
import { BACKEND_URL } from '../../config';
import { casdoorService } from '../identity/CasdoorService';

const { Paragraph, Text, Title } = Typography;
const ENTERPRISE_ADMIN_TABS = ['organization', 'members', 'invitations'] as const;
type EnterpriseAdminTab = typeof ENTERPRISE_ADMIN_TABS[number];

interface EnterpriseSummary {
  enterpriseId: string;
  enterpriseName: string;
  myRole: string;
  memberCount: number;
  adminCount: number;
  orgUnitCount: number;
  pendingInviteCount: number;
}

interface OrgUnit {
  id: string;
  parentId?: string;
  name: string;
  code?: string;
  managerUserId?: string;
  sortOrder: number;
  status: string;
}

interface MemberItem {
  userId: string;
  email: string;
  displayName: string;
  role: string;
  status: string;
  joinedAt: string;
  orgUnitId?: string;
  orgUnitName?: string;
  lastLoginAt?: string;
  sourceOrganizationId?: string;
  avatar?: string;
}

interface ExistingUser {
  userId: string;
  email: string;
  displayName: string;
}

interface InvitationItem {
  id: string;
  email: string;
  role: string;
  status: string;
  defaultOrgUnitId?: string;
  maxUses: number;
  usedCount: number;
  expiresAt?: string;
  inviteUrl?: string;
  createdAt: string;
}

function getAuthHeaders() {
  const token = casdoorService.getToken();
  return token ? { Authorization: `Bearer ${token}` } : {};
}

function formatDateTime(value?: string, locale = 'zh-CN') {
  if (!value) {
    return '-';
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }
  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date);
}

function buildOrgTree(units: OrgUnit[], parentId?: string): Array<{ key: string; title: React.ReactNode; children?: any[] }> {
  return units
    .filter((unit) => (unit.parentId || '') === (parentId || ''))
    .sort((a, b) => a.sortOrder - b.sortOrder || a.name.localeCompare(b.name))
    .map((unit) => ({
      key: unit.id,
      title: (
        <Space size={8}>
          <Text strong>{unit.name}</Text>
          {unit.code ? <Text type="secondary">{unit.code}</Text> : null}
        </Space>
      ),
      children: buildOrgTree(units, unit.id),
    }));
}

const AdminSettings: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [searchParams, setSearchParams] = useSearchParams();
  const [messageApi, contextHolder] = message.useMessage();
  const [fetching, setFetching] = useState(true);
  const [summary, setSummary] = useState<EnterpriseSummary | null>(null);
  const [orgUnits, setOrgUnits] = useState<OrgUnit[]>([]);
  const [members, setMembers] = useState<MemberItem[]>([]);
  const [invitations, setInvitations] = useState<InvitationItem[]>([]);
  const [enterpriseAdminDenied, setEnterpriseAdminDenied] = useState(false);
  const [createEnterpriseOpen, setCreateEnterpriseOpen] = useState(false);
  const [orgModalOpen, setOrgModalOpen] = useState(false);
  const [editingOrgUnit, setEditingOrgUnit] = useState<OrgUnit | null>(null);
  const [memberModalOpen, setMemberModalOpen] = useState(false);
  const [invitationModalOpen, setInvitationModalOpen] = useState(false);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [userSearchOptions, setUserSearchOptions] = useState<ExistingUser[]>([]);
  const [savingOrgUnit, setSavingOrgUnit] = useState(false);
  const [addingMember, setAddingMember] = useState(false);
  const [creatingInvitation, setCreatingInvitation] = useState(false);
  const [createEnterpriseForm] = Form.useForm<{ name: string }>();
  const [orgForm] = Form.useForm<OrgUnit>();
  const [memberForm] = Form.useForm<{ userId?: string; email?: string; role: string; orgUnitId?: string }>();
  const [invitationForm] = Form.useForm<{ email?: string; role: string; defaultOrgUnitId?: string; expiresInDays?: number; maxUses?: number }>();

  const activeTab = useMemo<EnterpriseAdminTab>(() => {
    const requestedTab = searchParams.get('tab');
    return ENTERPRISE_ADMIN_TABS.includes(requestedTab as EnterpriseAdminTab)
      ? (requestedTab as EnterpriseAdminTab)
      : 'organization';
  }, [searchParams]);

  const orgTreeData = useMemo(() => buildOrgTree(orgUnits), [orgUnits]);
  const orgUnitOptions = useMemo(() => orgUnits.map((unit) => ({
    label: unit.name,
    value: unit.id,
  })), [orgUnits]);
  const orgUnitMap = useMemo(() => new Map(orgUnits.map((unit) => [unit.id, unit])), [orgUnits]);
  const activeTabAction = useMemo(() => {
    if (activeTab === 'organization') {
      return {
        label: t('enterprise_admin_new_department'),
        icon: <PlusOutlined />,
        onClick: () => {
          setEditingOrgUnit(null);
          orgForm.setFieldsValue({ name: '', code: '', parentId: undefined, managerUserId: undefined, sortOrder: 100 });
          setOrgModalOpen(true);
        },
      };
    }
    if (activeTab === 'members') {
      return {
        label: t('enterprise_admin_add_existing_member'),
        icon: <PlusOutlined />,
        onClick: () => setMemberModalOpen(true),
      };
    }
    return {
      label: t('enterprise_admin_create_invite'),
      icon: <PlusOutlined />,
      onClick: () => setInvitationModalOpen(true),
    };
  }, [activeTab, orgForm, t]);

  const loadEnterpriseData = async () => {
    try {
      const [summaryRes, orgUnitsRes, membersRes, invitationsRes] = await Promise.all([
        axios.get(`${BACKEND_URL}/api/admin/summary`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/org-units`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/members`, { headers: getAuthHeaders() }),
        axios.get(`${BACKEND_URL}/api/admin/invitations`, { headers: getAuthHeaders() }),
      ]);
      setSummary(summaryRes.data || null);
      setOrgUnits(Array.isArray(orgUnitsRes.data) ? orgUnitsRes.data : []);
      setMembers(Array.isArray(membersRes.data) ? membersRes.data : []);
      setInvitations(Array.isArray(invitationsRes.data) ? invitationsRes.data : []);
      setEnterpriseAdminDenied(false);
    } catch (error: any) {
      if (error?.response?.status === 403) {
        setEnterpriseAdminDenied(true);
        setSummary(null);
        setOrgUnits([]);
        setMembers([]);
        setInvitations([]);
        return;
      }
      throw error;
    }
  };

  const loadAll = async () => {
    setFetching(true);
    try {
      await loadEnterpriseData();
    } catch {
      messageApi.error(t('enterprise_admin_load_failed'));
    } finally {
      setFetching(false);
    }
  };

  useEffect(() => {
    loadAll();
  }, [t]);

  const handleTabChange = (nextTab: string) => {
    const tab = ENTERPRISE_ADMIN_TABS.includes(nextTab as EnterpriseAdminTab)
      ? (nextTab as EnterpriseAdminTab)
      : 'organization';
    const nextSearchParams = new URLSearchParams(searchParams);
    nextSearchParams.set('tab', tab);
    setSearchParams(nextSearchParams, { replace: true });
  };

  const handleCreateEnterprise = async () => {
    const values = await createEnterpriseForm.validateFields();
    try {
      await axios.post(`${BACKEND_URL}/api/enterprises`, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_enterprise_created'));
      setCreateEnterpriseOpen(false);
      createEnterpriseForm.resetFields();
      window.location.reload();
    } catch {
      messageApi.error(t('enterprise_admin_enterprise_create_failed'));
    }
  };

  const openCreateOrgUnit = () => {
    setEditingOrgUnit(null);
    orgForm.setFieldsValue({ name: '', code: '', parentId: undefined, managerUserId: undefined, sortOrder: 100 });
    setOrgModalOpen(true);
  };

  const openEditOrgUnit = (unit: OrgUnit) => {
    setEditingOrgUnit(unit);
    orgForm.setFieldsValue(unit);
    setOrgModalOpen(true);
  };

  const handleSaveOrgUnit = async () => {
    const values = await orgForm.validateFields();
    setSavingOrgUnit(true);
    try {
      if (editingOrgUnit) {
        await axios.put(`${BACKEND_URL}/api/admin/org-units/${editingOrgUnit.id}`, values, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('enterprise_admin_department_updated'));
      } else {
        await axios.post(`${BACKEND_URL}/api/admin/org-units`, values, {
          headers: getAuthHeaders(),
        });
        messageApi.success(t('enterprise_admin_department_created'));
      }
      setOrgModalOpen(false);
      orgForm.resetFields();
      await loadEnterpriseData();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('enterprise_admin_department_save_failed'));
    } finally {
      setSavingOrgUnit(false);
    }
  };

  const handleDeleteOrgUnit = async (id: string) => {
    try {
      await axios.delete(`${BACKEND_URL}/api/admin/org-units/${id}`, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_department_deleted'));
      await loadEnterpriseData();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('enterprise_admin_department_delete_failed'));
    }
  };

  const handleSearchUsers = async (query: string) => {
    if (!query.trim()) {
      setUserSearchOptions([]);
      return;
    }
    setUserSearchLoading(true);
    try {
      const res = await axios.get(`${BACKEND_URL}/api/admin/users/search`, {
        headers: getAuthHeaders(),
        params: { query: query.trim() },
      });
      setUserSearchOptions(Array.isArray(res.data) ? res.data : []);
    } catch {
      setUserSearchOptions([]);
    } finally {
      setUserSearchLoading(false);
    }
  };

  const handleAddExistingMember = async () => {
    const values = await memberForm.validateFields();
    setAddingMember(true);
    try {
      const selectedUser = userSearchOptions.find((item) => item.userId === values.userId);
      await axios.post(`${BACKEND_URL}/api/admin/members/add-existing`, {
        userId: values.userId,
        email: selectedUser?.email || values.email,
        role: values.role,
        orgUnitId: values.orgUnitId,
      }, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_member_added'));
      setMemberModalOpen(false);
      memberForm.resetFields();
      setUserSearchOptions([]);
      await loadEnterpriseData();
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('enterprise_admin_member_add_failed'));
    } finally {
      setAddingMember(false);
    }
  };

  const handleUpdateMemberRole = async (userId: string, role: string) => {
    try {
      await axios.put(`${BACKEND_URL}/api/admin/members/${userId}/role`, { role }, {
        headers: getAuthHeaders(),
      });
      setMembers((current) => current.map((item) => (item.userId === userId ? { ...item, role } : item)));
      messageApi.success(t('enterprise_admin_role_updated'));
    } catch {
      messageApi.error(t('enterprise_admin_role_update_failed'));
    }
  };

  const handleUpdateMemberOrgUnit = async (userId: string, orgUnitId?: string) => {
    try {
      await axios.put(`${BACKEND_URL}/api/admin/members/${userId}/org-unit`, { orgUnitId }, {
        headers: getAuthHeaders(),
      });
      setMembers((current) => current.map((item) => (
        item.userId === userId
          ? { ...item, orgUnitId, orgUnitName: orgUnitId ? orgUnitMap.get(orgUnitId)?.name : undefined }
          : item
      )));
      messageApi.success(t('enterprise_admin_member_department_updated'));
    } catch {
      messageApi.error(t('enterprise_admin_member_department_update_failed'));
    }
  };

  const handleCreateInvitation = async () => {
    const values = await invitationForm.validateFields();
    setCreatingInvitation(true);
    try {
      const res = await axios.post(`${BACKEND_URL}/api/admin/invitations`, values, {
        headers: getAuthHeaders(),
      });
      messageApi.success(t('enterprise_admin_invitation_created'));
      setInvitationModalOpen(false);
      invitationForm.resetFields();
      await loadEnterpriseData();
      if (res.data?.inviteUrl && navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(res.data.inviteUrl);
        messageApi.success(t('enterprise_admin_invitation_link_copied'));
      }
    } catch (error: any) {
      const errorText = error?.response?.data;
      messageApi.error(typeof errorText === 'string' ? errorText : t('enterprise_admin_invitation_create_failed'));
    } finally {
      setCreatingInvitation(false);
    }
  };

  const memberColumns = [
    {
      title: t('enterprise_admin_members_column_member'),
      dataIndex: 'displayName',
      key: 'displayName',
      render: (_: string, item: MemberItem) => (
        <Space orientation="vertical" size={0}>
          <Text strong>{item.displayName || item.email || item.userId}</Text>
          <Text type="secondary">{item.email || item.userId}</Text>
        </Space>
      ),
    },
    {
      title: t('enterprise_admin_members_column_role'),
      dataIndex: 'role',
      key: 'role',
      width: 160,
      render: (value: string, item: MemberItem) => (
        <Select
          value={value}
          style={{ width: '100%' }}
          options={[
            { label: t('enterprise_role_owner'), value: 'owner' },
            { label: t('enterprise_role_admin'), value: 'admin' },
            { label: t('enterprise_role_member'), value: 'member' },
          ]}
          onChange={(next) => handleUpdateMemberRole(item.userId, next)}
        />
      ),
    },
    {
      title: t('enterprise_admin_members_column_department'),
      dataIndex: 'orgUnitName',
      key: 'orgUnitName',
      width: 220,
      render: (_: string, item: MemberItem) => (
        <Select
          allowClear
          value={item.orgUnitId}
          placeholder={t('enterprise_admin_select_department')}
          style={{ width: '100%' }}
          options={orgUnitOptions}
          onChange={(next) => handleUpdateMemberOrgUnit(item.userId, next)}
        />
      ),
    },
    {
      title: t('enterprise_admin_members_column_joined'),
      dataIndex: 'joinedAt',
      key: 'joinedAt',
      width: 180,
      render: (value: string) => formatDateTime(value, i18n.language),
    },
    {
      title: t('enterprise_admin_members_column_last_login'),
      dataIndex: 'lastLoginAt',
      key: 'lastLoginAt',
      width: 180,
      render: (value?: string) => formatDateTime(value, i18n.language),
    },
  ];

  const invitationColumns = [
    {
      title: t('enterprise_admin_invites_column_target'),
      dataIndex: 'email',
      key: 'email',
      render: (value: string) => value || <Text type="secondary">{t('enterprise_admin_open_invite_link')}</Text>,
    },
    {
      title: t('enterprise_admin_invites_column_role'),
      dataIndex: 'role',
      key: 'role',
      width: 120,
      render: (value: string) => (
        <Tag color={value === 'owner' ? 'gold' : value === 'admin' ? 'blue' : 'default'}>
          {value === 'owner' ? t('enterprise_role_owner') : value === 'admin' ? t('enterprise_role_admin') : t('enterprise_role_member')}
        </Tag>
      ),
    },
    {
      title: t('enterprise_admin_invites_column_status'),
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value: string) => (
        <Tag color={value === 'pending' ? 'processing' : value === 'accepted' ? 'success' : 'default'}>
          {value === 'pending' ? t('enterprise_invite_status_pending') : value === 'accepted' ? t('enterprise_invite_status_accepted') : value}
        </Tag>
      ),
    },
    {
      title: t('enterprise_admin_invites_column_usage'),
      key: 'usage',
      width: 100,
      render: (_: unknown, item: InvitationItem) => `${item.usedCount}/${item.maxUses}`,
    },
    {
      title: t('enterprise_admin_invites_column_expires'),
      dataIndex: 'expiresAt',
      key: 'expiresAt',
      width: 180,
      render: (value?: string) => formatDateTime(value, i18n.language),
    },
    {
      title: t('enterprise_admin_invites_column_link'),
      dataIndex: 'inviteUrl',
      key: 'inviteUrl',
      render: (value?: string) => (
        <Button
          icon={<CopyOutlined />}
          disabled={!value}
          onClick={async () => {
            if (!value) {
              return;
            }
            await navigator.clipboard.writeText(value);
            messageApi.success(t('enterprise_admin_invitation_link_copied'));
          }}
        >
          {t('enterprise_admin_copy_link')}
        </Button>
      ),
    },
  ];

  return (
    <div style={{ animation: 'fadeIn 0.5s ease-out' }}>
      {contextHolder}
      <Space orientation="vertical" size={24} style={{ width: '100%' }}>
        {enterpriseAdminDenied ? (
          <Card variant="borderless" loading={fetching} style={{ borderRadius: 16 }}>
            <Empty
              description={t('enterprise_admin_denied')}
              image={Empty.PRESENTED_IMAGE_SIMPLE}
            />
          </Card>
        ) : (
          <>
            <Card
              variant="borderless"
              loading={fetching}
              style={{
                borderRadius: 24,
                overflow: 'hidden',
                background: 'linear-gradient(135deg, #f8fbff 0%, #eef6ff 42%, #f9fcff 100%)',
                border: '1px solid #d6e4ff',
              }}
              styles={{ body: { padding: 24 } }}
            >
              <Row gutter={[24, 24]} align="middle">
                <Col xs={24} xl={16}>
                  <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                    <Tag color="blue" style={{ width: 'fit-content', borderRadius: 999, paddingInline: 12 }}>
                      {t('enterprise_admin_current_enterprise')}
                    </Tag>
                    <div>
                      <Title level={3} style={{ marginBottom: 8 }}>
                        <SettingOutlined style={{ marginRight: 10 }} />
                        {summary?.enterpriseName || t('enterprise_admin_title')}
                      </Title>
                      <Paragraph type="secondary" style={{ marginBottom: 0, maxWidth: 720 }}>
                        {t('enterprise_admin_summary_desc')}
                      </Paragraph>
                    </div>
                    <Space wrap size={[12, 12]}>
                      <Tag color="geekblue" style={{ borderRadius: 999, paddingInline: 12 }}>
                        {t('enterprise_admin_current_enterprise')}: {summary?.enterpriseName || '-'}
                      </Tag>
                      <Tag color="gold" style={{ borderRadius: 999, paddingInline: 12 }}>
                        {t('enterprise_admin_my_role', {
                          role: summary?.myRole ? t(`enterprise_role_${summary.myRole}`) : '-',
                        })}
                      </Tag>
                    </Space>
                  </Space>
                </Col>
                <Col xs={24} xl={8}>
                  <Space orientation="vertical" size={12} style={{ width: '100%' }}>
                    <Button type="primary" size="large" icon={<PlusOutlined />} onClick={() => setCreateEnterpriseOpen(true)} block>
                      {t('enterprise_admin_create_enterprise')}
                    </Button>
                    <Button size="large" icon={activeTabAction.icon} onClick={activeTabAction.onClick} block>
                      {activeTabAction.label}
                    </Button>
                  </Space>
                </Col>
              </Row>
            </Card>

            <Row gutter={[16, 16]}>
              <Col xs={24} sm={12} lg={6}>
                <Card variant="borderless" loading={fetching} style={{ borderRadius: 20, height: '100%' }}>
                  <Statistic title={t('enterprise_admin_current_enterprise')} value={summary?.enterpriseName || '-'} prefix={<TeamOutlined />} />
                </Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card variant="borderless" loading={fetching} style={{ borderRadius: 20, height: '100%' }}>
                  <Statistic title={t('enterprise_admin_members')} value={summary?.memberCount || 0} />
                </Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card variant="borderless" loading={fetching} style={{ borderRadius: 20, height: '100%' }}>
                  <Statistic title={t('enterprise_admin_departments')} value={summary?.orgUnitCount || 0} prefix={<ApartmentOutlined />} />
                </Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card variant="borderless" loading={fetching} style={{ borderRadius: 20, height: '100%' }}>
                  <Statistic title={t('enterprise_admin_admins')} value={summary?.adminCount || 0} />
                </Card>
              </Col>
              <Col xs={12} sm={12} lg={6}>
                <Card variant="borderless" loading={fetching} style={{ borderRadius: 20, height: '100%' }}>
                  <Statistic title={t('enterprise_admin_pending_invites')} value={summary?.pendingInviteCount || 0} />
                </Card>
              </Col>
            </Row>

            <Tabs
              activeKey={activeTab}
              onChange={handleTabChange}
              size="large"
              items={[
                {
                  key: 'organization',
                  label: t('enterprise_admin_tab_organization'),
                  children: (
                    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                      <Card variant="borderless" style={{ borderRadius: 20 }}>
                        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                          {t('enterprise_admin_organization_desc')}
                        </Paragraph>
                      </Card>
                      <Row gutter={[16, 16]}>
                        <Col xs={24} xl={10}>
                          <Card
                            variant="borderless"
                            title={<Space><ApartmentOutlined />{t('enterprise_admin_org_tree')}</Space>}
                            extra={<Button type="primary" icon={<PlusOutlined />} onClick={openCreateOrgUnit}>{t('enterprise_admin_new_department')}</Button>}
                            style={{ borderRadius: 20, minHeight: 480 }}
                          >
                            {orgTreeData.length ? (
                              <Tree defaultExpandAll treeData={orgTreeData} />
                            ) : (
                              <Empty description={t('enterprise_admin_no_departments')} image={Empty.PRESENTED_IMAGE_SIMPLE} />
                            )}
                          </Card>
                        </Col>
                        <Col xs={24} xl={14}>
                          <Card variant="borderless" title={t('enterprise_admin_department_directory')} style={{ borderRadius: 20, minHeight: 480 }}>
                            <Table
                              rowKey="id"
                              loading={fetching}
                              pagination={false}
                              dataSource={orgUnits}
                              columns={[
                                {
                                  title: t('enterprise_admin_department_name'),
                                  dataIndex: 'name',
                                  key: 'name',
                                  render: (_: string, item: OrgUnit) => (
                                    <div style={{ display: 'flex', flexDirection: 'column' }}>
                                      <Text strong>{item.name}</Text>
                                      <Text type="secondary">{item.code || t('enterprise_admin_no_code')}</Text>
                                    </div>
                                  ),
                                },
                                {
                                  title: t('enterprise_admin_department_parent'),
                                  dataIndex: 'parentId',
                                  key: 'parentId',
                                  render: (value?: string) => value ? orgUnitMap.get(value)?.name || '-' : <Tag>{t('enterprise_admin_root_department')}</Tag>,
                                },
                                {
                                  title: t('enterprise_admin_department_sort'),
                                  dataIndex: 'sortOrder',
                                  key: 'sortOrder',
                                  width: 90,
                                },
                                {
                                  title: t('enterprise_admin_actions'),
                                  key: 'actions',
                                  width: 180,
                                  render: (_: unknown, item: OrgUnit) => (
                                    <Space>
                                      <Button onClick={() => openEditOrgUnit(item)}>{t('enterprise_admin_edit')}</Button>
                                      <Button danger onClick={() => handleDeleteOrgUnit(item.id)}>{t('enterprise_admin_delete')}</Button>
                                    </Space>
                                  ),
                                },
                              ]}
                            />
                          </Card>
                        </Col>
                      </Row>
                    </Space>
                  ),
                },
                {
                  key: 'members',
                  label: t('enterprise_admin_tab_members'),
                  children: (
                    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                      <Card variant="borderless" style={{ borderRadius: 20 }}>
                        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                          {t('enterprise_admin_members_desc')}
                        </Paragraph>
                      </Card>
                      <Card
                        variant="borderless"
                        title={<Space><TeamOutlined />{t('enterprise_admin_members_title')}</Space>}
                        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setMemberModalOpen(true)}>{t('enterprise_admin_add_existing_member')}</Button>}
                        style={{ borderRadius: 20 }}
                      >
                        <Table rowKey="userId" loading={fetching} dataSource={members} columns={memberColumns} scroll={{ x: 980 }} />
                      </Card>
                    </Space>
                  ),
                },
                {
                  key: 'invitations',
                  label: t('enterprise_admin_tab_invitations'),
                  children: (
                    <Space orientation="vertical" size={16} style={{ width: '100%' }}>
                      <Card variant="borderless" style={{ borderRadius: 20 }}>
                        <Paragraph type="secondary" style={{ marginBottom: 0 }}>
                          {t('enterprise_admin_invitations_desc')}
                        </Paragraph>
                      </Card>
                      <Card
                        variant="borderless"
                        title={<Space><LinkOutlined />{t('enterprise_admin_invitations_title')}</Space>}
                        extra={<Button type="primary" icon={<PlusOutlined />} onClick={() => setInvitationModalOpen(true)}>{t('enterprise_admin_create_invite')}</Button>}
                        style={{ borderRadius: 20 }}
                      >
                        <Table rowKey="id" loading={fetching} dataSource={invitations} columns={invitationColumns} scroll={{ x: 920 }} />
                      </Card>
                    </Space>
                  ),
                },
              ]}
            />
          </>
        )}
      </Space>

      <Modal
        title={t('enterprise_admin_create_enterprise')}
        open={createEnterpriseOpen}
        forceRender
        onOk={handleCreateEnterprise}
        onCancel={() => setCreateEnterpriseOpen(false)}
        okText={t('enterprise_admin_create')}
      >
        <Form form={createEnterpriseForm} layout="vertical">
          <Form.Item label={t('enterprise_admin_enterprise_name')} name="name" rules={[{ required: true, message: t('enterprise_admin_enterprise_name_required') }]}>
            <Input placeholder={t('enterprise_admin_enterprise_name_placeholder')} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={editingOrgUnit ? t('enterprise_admin_edit_department') : t('enterprise_admin_create_department')}
        open={orgModalOpen}
        forceRender
        onOk={handleSaveOrgUnit}
        onCancel={() => setOrgModalOpen(false)}
        confirmLoading={savingOrgUnit}
        okText={editingOrgUnit ? t('enterprise_admin_save') : t('enterprise_admin_create')}
      >
        <Form form={orgForm} layout="vertical">
          <Form.Item label={t('enterprise_admin_department_name')} name="name" rules={[{ required: true, message: t('enterprise_admin_department_name_required') }]}>
            <Input placeholder={t('enterprise_admin_department_name_placeholder')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_department_code')} name="code">
            <Input placeholder={t('enterprise_admin_department_code_placeholder')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_department_parent')} name="parentId">
            <Select allowClear options={orgUnitOptions} placeholder={t('enterprise_admin_root_department')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_manager_user_id')} name="managerUserId">
            <Input placeholder={t('enterprise_admin_optional')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_department_sort')} name="sortOrder">
            <InputNumber style={{ width: '100%' }} min={0} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('enterprise_admin_add_existing_member')}
        open={memberModalOpen}
        forceRender
        onOk={handleAddExistingMember}
        onCancel={() => setMemberModalOpen(false)}
        confirmLoading={addingMember}
        okText={t('enterprise_admin_add_member')}
      >
        <Form form={memberForm} layout="vertical" initialValues={{ role: 'member' }}>
          <Form.Item
            label={t('enterprise_admin_search_user')}
            name="userId"
            rules={[{ required: true, message: t('enterprise_admin_user_required') }]}
          >
            <Select
              showSearch
              placeholder={t('enterprise_admin_search_user_placeholder')}
              filterOption={false}
              loading={userSearchLoading}
              onSearch={handleSearchUsers}
              options={userSearchOptions.map((item) => ({
                label: `${item.displayName || item.email} (${item.email || item.userId})`,
                value: item.userId,
              }))}
            />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_members_column_role')} name="role" rules={[{ required: true, message: t('enterprise_admin_role_required') }]}>
            <Select
              options={[
                { label: t('enterprise_role_owner'), value: 'owner' },
                { label: t('enterprise_role_admin'), value: 'admin' },
                { label: t('enterprise_role_member'), value: 'member' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_members_column_department')} name="orgUnitId">
            <Select allowClear options={orgUnitOptions} placeholder={t('enterprise_admin_optional')} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={t('enterprise_admin_create_invite')}
        open={invitationModalOpen}
        forceRender
        onOk={handleCreateInvitation}
        onCancel={() => setInvitationModalOpen(false)}
        confirmLoading={creatingInvitation}
        okText={t('enterprise_admin_create')}
      >
        <Form form={invitationForm} layout="vertical" initialValues={{ role: 'member', expiresInDays: 7, maxUses: 1 }}>
          <Form.Item label={t('enterprise_admin_restricted_email')} name="email">
            <Input placeholder={t('enterprise_admin_restricted_email_placeholder')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_members_column_role')} name="role" rules={[{ required: true, message: t('enterprise_admin_role_required') }]}>
            <Select
              options={[
                { label: t('enterprise_role_owner'), value: 'owner' },
                { label: t('enterprise_role_admin'), value: 'admin' },
                { label: t('enterprise_role_member'), value: 'member' },
              ]}
            />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_default_department')} name="defaultOrgUnitId">
            <Select allowClear options={orgUnitOptions} placeholder={t('enterprise_admin_optional')} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_expires_in_days')} name="expiresInDays">
            <InputNumber style={{ width: '100%' }} min={1} max={90} />
          </Form.Item>
          <Form.Item label={t('enterprise_admin_max_uses')} name="maxUses">
            <InputNumber style={{ width: '100%' }} min={1} max={50} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default AdminSettings;
