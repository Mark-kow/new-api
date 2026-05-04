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

export default function SettingsPaymentGatewayWechat(props) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WechatPayEnabled: false,
    WechatPayAppID: '',
    WechatPayMchID: '',
    WechatPaySerialNo: '',
    WechatPayAPIv3Key: '',
    WechatPayPrivateKey: '',
    WechatPayPlatformCert: '',
    WechatPayNotifyURL: '',
    WechatPayH5Domain: '',
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WechatPayEnabled:
          props.options.WechatPayEnabled === true ||
          props.options.WechatPayEnabled === 'true',
        WechatPayAppID: props.options.WechatPayAppID || '',
        WechatPayMchID: props.options.WechatPayMchID || '',
        WechatPaySerialNo: props.options.WechatPaySerialNo || '',
        WechatPayAPIv3Key: props.options.WechatPayAPIv3Key || '',
        WechatPayPrivateKey: props.options.WechatPayPrivateKey || '',
        WechatPayPlatformCert: props.options.WechatPayPlatformCert || '',
        WechatPayNotifyURL: props.options.WechatPayNotifyURL || '',
        WechatPayH5Domain: props.options.WechatPayH5Domain || '',
      };
      setInputs(currentInputs);
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
          key: 'WechatPayEnabled',
          value: inputs.WechatPayEnabled ? 'true' : 'false',
        },
        { key: 'WechatPayAppID', value: inputs.WechatPayAppID || '' },
        { key: 'WechatPayMchID', value: inputs.WechatPayMchID || '' },
        { key: 'WechatPaySerialNo', value: inputs.WechatPaySerialNo || '' },
        { key: 'WechatPayNotifyURL', value: inputs.WechatPayNotifyURL || '' },
        { key: 'WechatPayH5Domain', value: inputs.WechatPayH5Domain || '' },
        {
          key: 'WechatPayPlatformCert',
          value: inputs.WechatPayPlatformCert || '',
        },
      ];
      if (inputs.WechatPayAPIv3Key) {
        options.push({
          key: 'WechatPayAPIv3Key',
          value: inputs.WechatPayAPIv3Key,
        });
      }
      if (inputs.WechatPayPrivateKey) {
        options.push({
          key: 'WechatPayPrivateKey',
          value: inputs.WechatPayPrivateKey,
        });
      }

      const results = await Promise.all(
        options.map((opt) => API.put('/api/option/', opt)),
      );
      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => showError(res.data.message));
      } else {
        showSuccess(t('更新成功'));
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
        <Form.Section text={t('微信支付官方设置')}>
          <Text>{t('用于微信支付官方直连下单，支持 PC 扫码和移动 H5。')}</Text>
          <Banner
            type='info'
            description={`${t('异步通知地址')}：${props.options.ServerAddress || t('网站地址')}/api/wechat/notify`}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}>
            <Col xs={24} md={8}>
              <Form.Switch
                field='WechatPayEnabled'
                label={t('启用微信支付官方支付')}
                checkedText='｜'
                uncheckedText='〇'
              />
            </Col>
            <Col xs={24} md={8}>
              <Form.Input field='WechatPayAppID' label={t('AppID')} />
            </Col>
            <Col xs={24} md={8}>
              <Form.Input field='WechatPayMchID' label={t('商户号')} />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} md={8}>
              <Form.Input
                field='WechatPaySerialNo'
                label={t('商户证书序列号')}
              />
            </Col>
            <Col xs={24} md={8}>
              <Form.Input
                field='WechatPayAPIv3Key'
                label={t('APIv3 Key')}
                type='password'
              />
            </Col>
            <Col xs={24} md={8}>
              <Form.Input
                field='WechatPayH5Domain'
                label={t('H5 场景域名')}
                placeholder={t('例如：https://example.com')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} md={12}>
              <Form.TextArea
                field='WechatPayPrivateKey'
                label={t('商户私钥')}
                autosize={{ minRows: 6 }}
                placeholder={t(
                  '粘贴微信支付商户私钥 PEM 内容，敏感信息不会回显',
                )}
              />
            </Col>
            <Col xs={24} md={12}>
              <Form.TextArea
                field='WechatPayPlatformCert'
                label={t('微信支付平台证书')}
                autosize={{ minRows: 6 }}
                placeholder={t('粘贴微信支付平台证书 PEM 内容')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24}>
              <Form.Input
                field='WechatPayNotifyURL'
                label={t('自定义异步通知地址')}
                placeholder={t('留空则自动使用默认通知地址')}
              />
            </Col>
          </Row>
          <Button onClick={submit}>{t('更新微信支付设置')}</Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
