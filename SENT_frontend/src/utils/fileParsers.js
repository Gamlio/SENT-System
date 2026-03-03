import * as XLSX from 'xlsx';
import mammoth from 'mammoth';

// 1. Hàm đọc file Excel (.xlsx, .xls)
export const parseExcel = async (file) => {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        
        reader.onload = (e) => {
            try {
                const data = new Uint8Array(e.target.result);
                const workbook = XLSX.read(data, { type: 'array' });
                
                // Lấy sheet đầu tiên
                const firstSheetName = workbook.SheetNames[0];
                const worksheet = workbook.Sheets[firstSheetName];
                
                // Chuyển sheet thành mảng JSON (lấy cột A hoặc mảng 2 chiều)
                // header: 1 nghĩa là lấy dạng mảng mảng [[row1], [row2]]
                const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1 });
                
                // Làm phẳng mảng và lọc giá trị rỗng
                const flatData = jsonData
                    .flat()
                    .filter(item => item && String(item).trim() !== '');
                
                resolve(flatData.join('\n')); // Trả về chuỗi text ngăn cách bởi xuống dòng
            } catch (error) {
                reject("Lỗi định dạng Excel: " + error.message);
            }
        };
        
        reader.onerror = () => reject("Lỗi đọc file Excel");
        reader.readAsArrayBuffer(file);
    });
};

// 2. Hàm đọc file Word (.docx)
export const parseWord = async (file) => {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        
        reader.onload = async (e) => {
            try {
                const arrayBuffer = e.target.result;
                // mammoth giúp trích xuất text thô từ file word
                const result = await mammoth.extractRawText({ arrayBuffer: arrayBuffer });
                
                // Lọc các dòng trống thừa thãi
                const cleanText = result.value
                    .split('\n')
                    .map(line => line.trim())
                    .filter(line => line !== '')
                    .join('\n');
                    
                resolve(cleanText);
            } catch (error) {
                reject("Lỗi định dạng Word: " + error.message);
            }
        };
        
        reader.onerror = () => reject("Lỗi đọc file Word");
        reader.readAsArrayBuffer(file);
    });
};