// Módulo de generación de PDF en el navegador usando jsPDF (CDN)
import { formatearMoneda } from './ui.js';

/**
 * Genera y descarga un PDF de la factura dada usando la librería jsPDF.
 * @param {object} factura
 */
export function descargarFacturaPDF(factura) {
  // Verificar si jsPDF está cargado
  const { jsPDF } = window.jspdf || {};
  if (!jsPDF) {
    alert('La librería jsPDF no está disponible en este momento. Intente imprimir la página.');
    return;
  }

  const doc = new jsPDF({
    orientation: 'portrait',
    unit: 'mm',
    format: 'a4',
  });

  // Configuración de colores
  const azulPrimario = [43, 58, 74];
  const grisTexto = [80, 80, 80];

  // Encabezado
  doc.setFillColor(...azulPrimario);
  doc.rect(0, 0, 210, 25, 'F');

  doc.setTextColor(255, 255, 255);
  doc.setFontSize(18);
  doc.setFont('helvetica', 'bold');
  doc.text('SISTEMA DE FACTURACIÓN DIGITAL', 15, 16);

  // Información de Factura
  doc.setTextColor(...azulPrimario);
  doc.setFontSize(16);
  doc.text(`FACTURA Nº ${factura.factura_id}`, 15, 38);

  doc.setFontSize(10);
  doc.setFont('helvetica', 'normal');
  doc.setTextColor(...grisTexto);
  doc.text(`Fecha de emisión: ${factura.fecha}`, 15, 45);

  // Bloque Cliente y Usuario
  doc.setDrawColor(220, 224, 230);
  doc.setFillColor(248, 249, 250);
  doc.roundedRect(15, 52, 88, 38, 2, 2, 'FD');
  doc.roundedRect(107, 52, 88, 38, 2, 2, 'FD');

  // Datos Cliente
  doc.setTextColor(...azulPrimario);
  doc.setFont('helvetica', 'bold');
  doc.text('DATOS DEL CLIENTE', 20, 59);

  doc.setFont('helvetica', 'normal');
  doc.setTextColor(...grisTexto);
  doc.text(`Nombre: ${factura.cliente.nombre || ''}`, 20, 66);
  doc.text(`Cédula: ${factura.cliente.cc || ''}`, 20, 72);
  doc.text(`Email: ${factura.cliente.email || ''}`, 20, 78);
  doc.text(`Dirección: ${factura.cliente.direccion || ''}`, 20, 84);

  // Datos Facturador / Usuario
  doc.setTextColor(...azulPrimario);
  doc.setFont('helvetica', 'bold');
  doc.text('ATENDIDO POR', 112, 59);

  doc.setFont('helvetica', 'normal');
  doc.setTextColor(...grisTexto);
  doc.text(`Usuario: ${factura.usuario.nombre || ''}`, 112, 66);
  doc.text(`Rol: ${factura.usuario.rol || ''}`, 112, 72);
  doc.text(`ID Usuario: ${factura.usuario.id || ''}`, 112, 78);

  // Encabezado de la tabla de productos
  let y = 100;
  doc.setFillColor(...azulPrimario);
  doc.rect(15, y, 180, 8, 'F');

  doc.setTextColor(255, 255, 255);
  doc.setFont('helvetica', 'bold');
  doc.text('DESCRIPCIÓN', 20, y + 5.5);
  doc.text('CANT.', 105, y + 5.5, { align: 'center' });
  doc.text('PRECIO UNITARIO', 145, y + 5.5, { align: 'right' });
  doc.text('SUBTOTAL', 190, y + 5.5, { align: 'right' });

  y += 8;

  // Filas de productos
  doc.setFont('helvetica', 'normal');
  doc.setTextColor(...grisTexto);

  (factura.productos || []).forEach((prod, index) => {
    // Fondo alternado
    if (index % 2 === 1) {
      doc.setFillColor(248, 249, 250);
      doc.rect(15, y, 180, 7, 'F');
    }

    doc.text(prod.nombre || '', 20, y + 5);
    doc.text(String(prod.cantidad), 105, y + 5, { align: 'center' });
    doc.text(formatearMoneda(prod.precio_unitario), 145, y + 5, { align: 'right' });
    doc.text(formatearMoneda(prod.subtotal), 190, y + 5, { align: 'right' });

    y += 7;
  });

  // Línea separadora
  doc.setDrawColor(200, 200, 200);
  doc.line(15, y + 2, 195, y + 2);
  y += 8;

  // Total
  doc.setFillColor(240, 244, 248);
  doc.rect(115, y, 80, 12, 'F');

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(12);
  doc.setTextColor(...azulPrimario);
  doc.text('TOTAL:', 120, y + 8);
  doc.text(formatearMoneda(factura.total), 190, y + 8, { align: 'right' });

  // Pie de página
  doc.setFontSize(9);
  doc.setFont('helvetica', 'italic');
  doc.setTextColor(150, 150, 150);
  doc.text('Gracias por su compra. Documento generado por el Sistema de Microservicios.', 105, 280, { align: 'center' });

  // Guardar archivo
  doc.save(`factura-${factura.factura_id}.pdf`);
}
