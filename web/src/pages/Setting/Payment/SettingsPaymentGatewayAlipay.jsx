import React, { useEffect, useRef, useState } from 'react';
import {
  Banner,
  Button,
  Form,
  Row,
  Col,
  Typography,
  Spin,
} from '@douyinfe/semi-ui';
import { API, showError, showSuccess } from '../../../helpers';
import { useTranslation } from 'react-i18next';

const { Text } = Typography;

export default function SettingsPaymentGatewayAlipay(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayEnabled: false,
    AlipayAppID: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipayNotifyURL: '',
    AlipayReturnURL: '',
  });
  const [originInputs, setOriginInputs] = useState({});
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        AlipayEnabled:
          props.options.AlipayEnabled === true ||
          props.options.AlipayEnabled === 'true',
        AlipayAppID: props.options.AlipayAppID || '',
        AlipayPrivateKey: props.options.AlipayPrivateKey || '',
        AlipayPublicKey: props.options.AlipayPublicKey || '',
        AlipayNotifyURL: props.options.AlipayNotifyURL || '',
        AlipayReturnURL: props.options.AlipayReturnURL || '',
      };
      setInputs(currentInputs);
      setOriginInputs({ ...currentInputs });
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const submit = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      const options = [
        {
          key: 'AlipayEnabled',
          value: inputs.AlipayEnabled ? 'true' : 'false',
        },
        { key: 'AlipayAppID', value: inputs.AlipayAppID || '' },
        { key: 'AlipayNotifyURL', value: inputs.AlipayNotifyURL || '' },
        { key: 'AlipayReturnURL', value: inputs.AlipayReturnURL || '' },
      ];

      if (inputs.AlipayPrivateKey) {
        options.push({
          key: 'AlipayPrivateKey',
          value: inputs.AlipayPrivateKey,
        });
      }
      if (inputs.AlipayPublicKey) {
        options.push({ key: 'AlipayPublicKey', value: inputs.AlipayPublicKey });
      }

      const results = await Promise.all(
        options.map((opt) => API.put('/api/option/', opt)),
      );
      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => showError(res.data.message));
      } else {
        showSuccess(t('更新成功'));
        setOriginInputs({ ...inputs });
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={setInputs}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={t('支付宝官方设置')}>
          <Text>{t('用于支付宝官方直连下单，支持 PC 扫码和移动 H5。')}</Text>
          <Banner
            type='info'
            description={`${t('异步通知地址')}：${props.options.ServerAddress || t('网站地址')}/api/alipay/notify`}
          />
          <Banner
            type='info'
            description={`${t('同步回跳地址')}：${props.options.ServerAddress || t('网站地址')}/api/alipay/return`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}>
            <Col xs={24} md={8}>
              <Form.Switch
                field='AlipayEnabled'
                label={t('启用支付宝官方支付')}
                checkedText='｜'
                uncheckedText='〇'
              />
            </Col>
            <Col xs={24} md={16}>
              <Form.Input
                field='AlipayAppID'
                label={t('AppID')}
                placeholder={t('支付宝开放平台应用 AppID')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} md={12}>
              <Form.TextArea
                field='AlipayPrivateKey'
                label={t('应用私钥')}
                autosize={{ minRows: 6 }}
                placeholder={t('粘贴支付宝应用私钥 PEM 内容，敏感信息不会回显')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.TextArea
                field='AlipayPublicKey'
                label={t('支付宝公钥')}
                autosize={{ minRows: 6 }}
                placeholder={t('粘贴支付宝公钥 PEM 内容')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} md={12}>
              <Form.Input
                field='AlipayNotifyURL'
                label={t('自定义异步通知地址')}
                placeholder={t('留空则自动使用默认通知地址')}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.Input
                field='AlipayReturnURL'
                label={t('自定义同步回跳地址')}
                placeholder={t('留空则自动使用默认回跳地址')}
              />
            </Col>
          </Row>
          <Button onClick={submit}>{t('更新支付宝设置')}</Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
