import React from 'react';
import { Modal, Typography, Button } from '@douyinfe/semi-ui';
import { QRCodeSVG } from 'qrcode.react';

const { Text } = Typography;

const PaymentQRCodeModal = ({
  t,
  visible,
  onCancel,
  qrCodeUrl,
  tradeNo,
  paymentMethodName,
}) => {
  return (
    <Modal
      visible={visible}
      title={t('扫码支付')}
      onCancel={onCancel}
      footer={<Button onClick={onCancel}>{t('关闭')}</Button>}
      centered
      maskClosable={false}
    >
      <div className='flex flex-col items-center gap-4 py-2'>
        <Text>
          {t('请使用')} {paymentMethodName || t('支付应用')} {t('扫码完成支付')}
        </Text>
        {qrCodeUrl ? <QRCodeSVG value={qrCodeUrl} size={220} /> : null}
        <Text type='tertiary'>
          {t('订单号')}：{tradeNo}
        </Text>
        <Text type='tertiary'>{t('支付成功后将自动刷新状态')}</Text>
      </div>
    </Modal>
  );
};

export default PaymentQRCodeModal;
